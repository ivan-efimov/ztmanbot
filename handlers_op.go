package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
)

/* /op handler */
type OpHandler struct{}

func (OpHandler) Handle(msg *tgbotapi.Message, _ *ZeroTierApi, accessManager AccessManager) (tgbotapi.MessageConfig, error) {
	if accessManager.GetAccessLevel(msg.Chat.ID) < AccessLevelAdmin {
		return tgbotapi.NewMessage(msg.Chat.ID, AccessDeniedText), nil
	}

	args := splitArgs(msg.CommandArguments())
	if len(args) == 0 {
		return tgbotapi.NewMessage(msg.Chat.ID, "No arguments given. Try /help."), nil
	}
	if len(args) > 1 {
		return tgbotapi.NewMessage(msg.Chat.ID, "Too many arguments given. Try /help."), nil
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return tgbotapi.NewMessage(msg.Chat.ID, "Invalid argument. Try /help."), nil
	}

	err = accessManager.SetAccessLevel(id, AccessLevelOperator)
	if err != nil {
		if err == AdminMutationError {
			return tgbotapi.NewMessage(msg.Chat.ID, AccessDeniedText), nil
		}
		return tgbotapi.MessageConfig{}, err
	}
	return tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Success. %d is an operator now.", id)), nil
}

func (OpHandler) Description() string {
	return "Makes user with given user_id (number) an operator in app. Usage:`/op user_id`."
}

/* /deop handler */
type DeopHandler struct{}

func (DeopHandler) Handle(msg *tgbotapi.Message, _ *ZeroTierApi, accessManager AccessManager) (tgbotapi.MessageConfig, error) {
	if accessManager.GetAccessLevel(msg.Chat.ID) < AccessLevelAdmin {
		return tgbotapi.NewMessage(msg.Chat.ID, AccessDeniedText), nil
	}

	args := splitArgs(msg.CommandArguments())
	if len(args) == 0 {
		return tgbotapi.NewMessage(msg.Chat.ID, "No arguments given. Try /help."), nil
	}
	if len(args) > 1 {
		return tgbotapi.NewMessage(msg.Chat.ID, "Too many arguments given. Try /help."), nil
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return tgbotapi.NewMessage(msg.Chat.ID, "Invalid argument. Try /help."), nil
	}

	err = accessManager.SetAccessLevel(id, AccessLevelGuest)
	if err != nil {
		if err == AdminMutationError {
			return tgbotapi.NewMessage(msg.Chat.ID, AccessDeniedText), nil
		}
		return tgbotapi.MessageConfig{}, err
	}
	return tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Success. %d is a guest now.", id)), nil
}

func (DeopHandler) Description() string {
	return "Makes user with given user_id (number) a guest in app. Usage:`/deop user_id`."
}
