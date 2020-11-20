package main

import (
	"flag"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type WebhookConfigCustom struct {
	URL                *url.URL
	Certificate        interface{}
	MaxConnections     int
	DropPendingUpdates bool
	AllowedUpdates     []string
}

func SetWebhookCustom(bot *tgbotapi.BotAPI, config *WebhookConfigCustom) (tgbotapi.APIResponse, error) {
	if config.Certificate == nil {
		v := url.Values{}
		v.Add("url", config.URL.String())
		if config.DropPendingUpdates {
			v.Add("drop_pending_updates", "True")
		}
		if len(config.AllowedUpdates) > 0 {
			v.Add("allowed_updates", fmt.Sprint(config.AllowedUpdates))
		}
		if config.MaxConnections != 0 {
			v.Add("max_connections", strconv.Itoa(config.MaxConnections))
		}

		return bot.MakeRequest("setWebhook", v)
	}

	params := make(map[string]string)
	params["url"] = config.URL.String()
	if config.MaxConnections != 0 {
		params["max_connections"] = strconv.Itoa(config.MaxConnections)
	}
	if config.DropPendingUpdates {
		params["drop_pending_updates"] = "True"
	}
	if len(config.AllowedUpdates) > 0 {
		params["allowed_updates"] = fmt.Sprint(config.AllowedUpdates)
	}

	resp, err := bot.UploadFile("setWebhook", params, "certificate", config.Certificate)
	if err != nil {
		return tgbotapi.APIResponse{}, err
	}

	return resp, nil
}

func main() {
	configFile := flag.String("config", "", "path to a config file")
	debugMode := flag.Bool("debug", false, "run bot in debug mode")
	flag.Parse()

	botConfig, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading config: %s", err.Error())
	}

	accessManager, err := NewAccessManagerWithFileStorage(botConfig.AdminId, botConfig.OpsStorage)
	if err != nil {
		log.Fatalln(err)
	}

	ztApi := NewZTApi(botConfig.ZeroTierToken, botConfig.ZeroTierNetwork)

	commandManager := NewCommandManager(ztApi, accessManager)

	whURL, err := url.Parse(botConfig.WebHookUrl)
	if err != nil {
		log.Fatalln(err)
	}

	bot, err := tgbotapi.NewBotAPI(botConfig.Token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = *debugMode
	if bot.Debug {
		log.Println("Bot is running in DEBUG mode")
	}

	resp, err := SetWebhookCustom(bot, &WebhookConfigCustom{
		URL:                whURL,
		Certificate:        botConfig.WebHookCertFile,
		DropPendingUpdates: true,
		AllowedUpdates:     []string{"message"},
	})
	if err != nil {
		log.Fatalln(err)
	}
	if !resp.Ok {
		log.Fatalln(resp.Description)
	}
	defer bot.RemoveWebhook()

	updates := bot.ListenForWebhook(whURL.Path)
	go http.ListenAndServeTLS(botConfig.ListenAddr+":"+botConfig.ListenPort,
		botConfig.WebHookCertFile, botConfig.WebHookKeyFile, nil)

	for update := range updates {
		if update.Message == nil { // ignore all non-message updates
			continue
		}
		if update.Message.Chat.IsPrivate() {
			if *debugMode {
				log.Println("command:", update.Message.Command())
				log.Println("args:", update.Message.CommandArguments())
			}
			rep, err := commandManager.HandleMessage(update.Message)
			if err == nil {
				_, err = bot.Send(rep)
			} else {
				log.Println(err)
				errMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Something went wrong!")
				_, err = bot.Send(errMsg)
				if err != nil {
					log.Println(err)
				}
			}
		} else {
			errMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "I only work with private chats")
			_, err = bot.Send(errMsg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
