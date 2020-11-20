package main

import (
	"bytes"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"text/template"
)

/* /list handler */
type ListMembersHandler struct{}

func (ListMembersHandler) Handle(msg *tgbotapi.Message, ztApi *ZeroTierApi, _ AccessManager) (tgbotapi.MessageConfig, error) {
	args := splitArgs(msg.CommandArguments())
	if len(args) > 1 {
		return tgbotapi.NewMessage(msg.Chat.ID, "Too many arguments given. Try /help."), nil
	}
	type membersListCfg struct {
		Verbose bool
		Members []*MemberInfo
	}
	mList := &membersListCfg{false, nil}
	if len(args) == 1 {
		if args[0] == "-v" {
			mList.Verbose = true
		} else {
			return tgbotapi.NewMessage(msg.Chat.ID, "Invalid argument. Try /help."), nil
		}
	}
	members, err := ztApi.ListMembers(ztApi.defaultNetwork)
	if err != nil {
		return tgbotapi.MessageConfig{}, err
	}

	if members == nil {
		return tgbotapi.NewMessage(msg.Chat.ID,
			fmt.Sprintf("Failed to get members of %s.", ztApi.defaultNetwork)), nil
	}

	mList.Members = members

	var tStr = "{{$verbose := .Verbose}}" +
		"{{range $i, $member := .Members}}" +
		"{{$i}}.\nNodeID: {{$member.NodeID}}\n" +
		"Authorized: {{$member.Config.Authorized}}\n" +
		"Local Addresses:\n" +
		"{{range .Config.IpAssignments}}" +
		"> {{.}}\n" +
		"{{else}}" +
		"Not assigned\n" +
		"{{end}}" +
		"{{if $verbose}}" +
		"Name: {{$member.Name}}\n" +
		"Description: {{$member.Description}}\n" +
		"Hidden: {{$member.Hidden}}\n" +
		"Online: {{$member.Online}}\n" +
		"PhysicalAddress: {{$member.PhysicalAddress}}\n" +
		"ClientVersion: {{$member.ClientVersion}}\n" +
		"{{end}}" +
		"{{else}}" +
		"No members." +
		"{{end}}"

	var listTemplate template.Template
	_, err = listTemplate.Parse(tStr)
	if err != nil {
		return tgbotapi.MessageConfig{}, err
	}

	repBuf := bytes.NewBufferString("")
	err = listTemplate.Execute(repBuf, mList)
	if err != nil {
		return tgbotapi.MessageConfig{}, err
	}

	return tgbotapi.NewMessage(msg.Chat.ID, repBuf.String()), nil
}

func (ListMembersHandler) Description() string {
	return "Lists all nodes in network. Use -v if you want more details. Usage:`/list [-v]`."
}
