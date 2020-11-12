package main

import (
	"flag"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"net/url"
	"os"
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

	ztApi := NewZTApi(botConfig.ZeroTierToken, botConfig.ZeroTierNetwork)

	commandManager := NewCommandManager(ztApi)

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

	resp, err := bot.SetWebhook(tgbotapi.WebhookConfig{URL: whURL, Certificate: botConfig.WebHookCertFile})
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
