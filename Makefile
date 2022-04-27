.PHONY: build
build: bin/lanchat

.DEFAULT_GOAL: build

bin/lanchat: main.go config.go ui/*.go lan/*.go logger/*.go
	@CGO_ENABLED=1 go build -race -o ./bin/lanchat ./main.go ./config.go

.PHONY: test
test:
	go test -race ./...

.PHONY: check
check: bin/lanchat
	@./fake_chat.sh

.PHONY: test
test:
	@go test -race ./...
