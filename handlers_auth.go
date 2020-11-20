package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

/* /auth handler */
type AuthHandler struct{}

func (AuthHandler) Handle(msg *tgbotapi.Message, ztApi *ZeroTierApi, accessManager AccessManager) (tgbotapi.MessageConfig, error) {
	if accessManager.GetAccessLevel(msg.Chat.ID) < AccessLevelOperator {
		return tgbotapi.NewMessage(msg.Chat.ID, AccessDeniedText), nil
	}
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

func (UnauthHandler) Handle(msg *tgbotapi.Message, ztApi *ZeroTierApi, accessManager AccessManager) (tgbotapi.MessageConfig, error) {
	if accessManager.GetAccessLevel(msg.Chat.ID) < AccessLevelOperator {
		return tgbotapi.NewMessage(msg.Chat.ID, AccessDeniedText), nil
	}
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
