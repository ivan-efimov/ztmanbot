SOURCES=main.go

get_deps:
	go get -u github.com/go-telegram-bot-api/telegram-bot-api
build:
	go build -o zmanbot $(SOURCES)

all: get_deps build