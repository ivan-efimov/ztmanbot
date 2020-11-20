package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

const AccessDeniedText = "Access denied. If you think that's a mistake, contact you administrator."

type CommandHandler interface {
	Handle(*tgbotapi.Message, *ZeroTierApi, AccessManager) (tgbotapi.MessageConfig, error)
	Description() string
}

type CommandManager struct {
	registeredCommands map[string]CommandHandler
	ztApi              *ZeroTierApi
	accessManager      AccessManager
}

// Allocates new CommandManager with hardcoded registered commands
// If use want to implement new command you have create a handler type that implements CommandHandler interface
// and register it in this function the same way it done for already existing commands.
// I recommend to place the handler type in a separate file (look at `handlers_*.go` for example).
func NewCommandManager(ztApi *ZeroTierApi, accessManager AccessManager) *CommandManager {
	cm := &CommandManager{
		registeredCommands: make(map[string]CommandHandler),
		ztApi:              ztApi,
		accessManager:      accessManager,
	}
	cm.registeredCommands["start"] = StartHandler{}
	cm.registeredCommands["auth"] = AuthHandler{}
	cm.registeredCommands["unauth"] = UnauthHandler{}
	cm.registeredCommands["list"] = ListMembersHandler{}
	cm.registeredCommands["op"] = OpHandler{}
	cm.registeredCommands["deop"] = DeopHandler{}

	return cm
}

func (cm *CommandManager) HandleMessage(msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	if len(msg.Command()) == 0 {
		return tgbotapi.NewMessage(msg.Chat.ID, "I understand commands only. Try /help."), nil
	}
	if cm.accessManager.GetAccessLevel(msg.Chat.ID) < AccessLevelGuest {
		return tgbotapi.NewMessage(msg.Chat.ID, AccessDeniedText), nil
	}
	// help needs to be handled in special way
	if msg.Command() == "help" {
		if cm.accessManager.GetAccessLevel(msg.Chat.ID) < AccessLevelOperator {
			return tgbotapi.NewMessage(msg.Chat.ID, AccessDeniedText), nil
		}
		return tgbotapi.NewMessage(msg.Chat.ID, cm.HelpText()), nil
	}
	handler, found := cm.registeredCommands[msg.Command()]
	if !found {
		return tgbotapi.NewMessage(msg.Chat.ID, "Unknown command. Try /help."), nil
	}
	return handler.Handle(msg, cm.ztApi, cm.accessManager)
}

func (cm *CommandManager) HelpText() string {
	txt := "Help:\n" +
		"This bot is used to manage a ZeroTier network via ZeroTier-Central API.\n" +
		"Available commands:\n" +
		"/help : provides help.\n"
	for k, v := range cm.registeredCommands {
		txt += "/" + k + " : " + v.Description() + "\n"
	}
	return txt
}

func splitArgs(args string) []string {
	if len(args) == 0 {
		return nil
	}
	return strings.Split(args, " ")
}
