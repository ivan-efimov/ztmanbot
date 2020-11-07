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
	"regexp"
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

	if *configFile != "" {
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
	}

	ztApi := NewZTApi(botConfig.ZeroTierToken)

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

	addMemberRE := regexp.MustCompile("^[0-9a-f]{10}$")

	for update := range updates {
		if addMemberRE.MatchString(update.Message.Text) {
			ok, err := ztApi.AddMember(botConfig.ZeroTierNetwork, update.Message.Text)
			if err != nil {
				log.Println(err)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("Success: %t.\nNow you can join network %s.", ok, botConfig.ZeroTierNetwork))
			msg.ReplyToMessageID = update.Message.MessageID

			_, _ = bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Invalid input. Send node id (10 hexadecimal digits).")
			msg.ReplyToMessageID = update.Message.MessageID
			_, _ = bot.Send(msg)
		}
	}
}
