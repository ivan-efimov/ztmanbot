COM_HANDLERS=handlers_basic.go handlers_auth.go handlers_list.go handlers_op.go
SOURCES=main.go zerotierapi.go command.go config.go access_manager.go $(COM_HANDLERS)

get_deps:
	go get gopkg.in/yaml.v2
	go get -u github.com/go-telegram-bot-api/telegram-bot-api

build:
	go build -o zmanbot $(SOURCES)

fmt:
	gofmt -w $(SOURCES)

all: get_deps build