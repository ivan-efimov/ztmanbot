package main

import (
	"flag"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type BotConfig struct {
	Token           string `yaml:"token"`
	WebHookUrl      string `yaml:"web_hook_url"`
	WebHookCertFile string `yaml:"web_hook_cert"`
	WebHookKeyFile  string `yaml:"web_hook_key"`
	ListenAddr      string `yaml:"listen_addr"`
	ListenPort      string `yaml:"port"`
	ZeroTierToken   string `yaml:"zt_token"`
	ZeroTierNetwork string `yaml:"zt_network"`
}

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

	var botConfig BotConfig

	if len(*configFile) > 0 {
		cfg, err := os.Open(*configFile)
		if err != nil {
			log.Fatalf("Bad config file: %s\n", err.Error())
		}
		dec := yaml.NewDecoder(cfg)
		err = dec.Decode(&botConfig)
		_ = cfg.Close()
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Println("No config file given. Create config file such as following one (fill ALL empty strings):")
		sampleConfig := &BotConfig{}
		enc := yaml.NewEncoder(log.Writer())
		err := enc.Encode(sampleConfig)
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(1)
	}

	var accessManager AccessManager

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
