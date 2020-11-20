package main

import (
	"errors"
	"gopkg.in/yaml.v2"
	"log"
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
	AdminId         int64  `yaml:"admin_id"`
	OpsStorage      string `yaml:"ops_file"`
}

func LoadConfig(filename string) (BotConfig, error) {
	var botConfig BotConfig
	if len(filename) > 0 {
		cfg, err := os.Open(filename)
		if err != nil {
			return BotConfig{}, err
		}
		dec := yaml.NewDecoder(cfg)
		err = dec.Decode(&botConfig)
		_ = cfg.Close()
		if err != nil {
			return BotConfig{}, err
		}
	} else {
		log.Println("No config file given. Create config file such as following one (fill ALL empty strings):")
		sampleConfig := &BotConfig{}
		enc := yaml.NewEncoder(log.Writer())
		err := enc.Encode(sampleConfig)
		if err != nil {
			return BotConfig{}, err
		}
		return BotConfig{}, errors.New("no config file")
	}
	return botConfig, nil
}
