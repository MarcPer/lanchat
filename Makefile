.PHONY: build
build: lanchat

.DEFAULT_GOAL: build

bin/lanchat: main.go ui/*.go client/*.go
	@CGO_ENABLED=0 go build -race -o ./bin/lanchat ./main.go

.PHONY: test
test:
	go test -race ./...
