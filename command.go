package main

import (
	"bytes"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
	"text/template"
)

type CommandHandler interface {
	Handle(*tgbotapi.Message, *ZeroTierApi) (tgbotapi.MessageConfig, error)
	Description() string
}

type CommandManager struct {
	registeredCommands map[string]CommandHandler
	ztApi              *ZeroTierApi
}

func NewCommandManager(ztApi *ZeroTierApi) *CommandManager {
	cm := &CommandManager{
		registeredCommands: make(map[string]CommandHandler),
		ztApi:              ztApi,
	}
	cm.registeredCommands["start"] = StartHandler{}
	cm.registeredCommands["auth"] = AuthHandler{}
	cm.registeredCommands["unauth"] = UnauthHandler{}
	cm.registeredCommands["list"] = ListMembersHandler{}

	return cm
}

func (cm *CommandManager) HandleMessage(msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	if len(msg.Command()) == 0 {
		return tgbotapi.NewMessage(msg.Chat.ID, "I understand commands only. Try /help."), nil
	}
	// help needs to be handled in special way
	if msg.Command() == "help" {
		s := "Help:\nThis bot is used to manage a ZeroTier network via ZeroTier-Central API.\n"
		s += "Available commands:\n" + "/help : provides help.\n"
		for k, v := range cm.registeredCommands {
			s += "/" + k + " : " + v.Description() + "\n"
		}
		return tgbotapi.NewMessage(msg.Chat.ID, s), nil
	}
	handler, found := cm.registeredCommands[msg.Command()]
	if !found {
		return tgbotapi.NewMessage(msg.Chat.ID, "Unknown command. Try /help."), nil
	}
	return handler.Handle(msg, cm.ztApi)
}

/* /start handler */
type StartHandler struct{}

func (StartHandler) Handle(msg *tgbotapi.Message, ztApi *ZeroTierApi) (tgbotapi.MessageConfig, error) {
	return tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Hello, %d!", msg.Chat.ID)), nil
}

func (StartHandler) Description() string {
	return "begins interaction with me"
}

/* /auth handler */
type AuthHandler struct{}

func (AuthHandler) Handle(msg *tgbotapi.Message, ztApi *ZeroTierApi) (tgbotapi.MessageConfig, error) {
	args := splitArgs(msg.CommandArguments())
	if len(args) == 0 {
		return tgbotapi.NewMessage(msg.Chat.ID, "No arguments given. Try /help."), nil
	}
	if len(args) > 2 {
		return tgbotapi.NewMessage(msg.Chat.ID, "Too many arguments given. Try /help."), nil
	}
	nodeId := args[0]
	var shortname string
	if len(args) == 2 {
		shortname = args[1]
	}
	success, err := ztApi.AuthMember(
		ztApi.defaultNetwork, nodeId, shortname,
		fmt.Sprintf("added by via telegram bot by %d", msg.Chat.ID))
	if err != nil {
		if err == InvalidNodeId {
			return tgbotapi.NewMessage(msg.Chat.ID,
				"Invalid NodeID"), nil
		}
		return tgbotapi.MessageConfig{}, err
	}
	if success {
		return tgbotapi.NewMessage(msg.Chat.ID, "Success."), nil
	}
	return tgbotapi.NewMessage(msg.Chat.ID,
		fmt.Sprintf("Failed to authorize %s in %s!", ztApi.defaultNetwork, args[0])), nil

}

func (AuthHandler) Description() string {
	return "Authorizes given NodeID in network. Usage:`/auth NodeID short_name`."
}

/* /unauth handler */
type UnauthHandler struct{}

func (UnauthHandler) Handle(msg *tgbotapi.Message, ztApi *ZeroTierApi) (tgbotapi.MessageConfig, error) {
	args := splitArgs(msg.CommandArguments())
	if len(args) == 0 {
		return tgbotapi.NewMessage(msg.Chat.ID, "No arguments given. Try /help."), nil
	}
	if len(args) > 1 {
		return tgbotapi.NewMessage(msg.Chat.ID, "Too many arguments given. Try /help."), nil
	}
	success, err := ztApi.UnauthMemberByID(ztApi.defaultNetwork, args[0])
	if err != nil {
		if err == InvalidNodeId {
			return tgbotapi.NewMessage(msg.Chat.ID,
				"Invalid NodeID"), nil
		}
		return tgbotapi.MessageConfig{}, err
	}
	if success {
		return tgbotapi.NewMessage(msg.Chat.ID, "Success."), nil
	}
	return tgbotapi.NewMessage(msg.Chat.ID,
		fmt.Sprintf("Failed to unauthorize %s in %s!", ztApi.defaultNetwork, args[0])), nil

}

func (UnauthHandler) Description() string {
	return "Unauthorizes given NodeID in network. Usage:`/unauth NodeID`."
}

/* /list handler */
type ListMembersHandler struct{}

func (ListMembersHandler) Handle(msg *tgbotapi.Message, ztApi *ZeroTierApi) (tgbotapi.MessageConfig, error) {
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

func splitArgs(args string) []string {
	if len(args) == 0 {
		return nil
	}
	return strings.Split(args, " ")
}
