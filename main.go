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
		defer cfg.Close()
		dec := yaml.NewDecoder(cfg)
		err = dec.Decode(&botConfig)
		if err != nil {
			log.Fatalln(err)
		}
	}

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
		log.Printf("%+v\n", update)
	}
}
