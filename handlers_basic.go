package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

/* /start handler */
type StartHandler struct{}

func (StartHandler) Handle(msg *tgbotapi.Message, _ *ZeroTierApi, _ AccessManager) (tgbotapi.MessageConfig, error) {
	return tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Hello, %d!", msg.Chat.ID)), nil
}

func (StartHandler) Description() string {
	return "begins interaction with me"
}
