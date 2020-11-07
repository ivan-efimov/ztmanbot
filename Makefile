SOURCES=main.go zerotierapi.go

get_deps:
	go get gopkg.in/yaml.v2
	go get -u github.com/go-telegram-bot-api/telegram-bot-api

build:
	go build -o zmanbot $(SOURCES)

all: get_deps build