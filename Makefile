.PHONY: build

build: lanchat

lanchat: main.go client/client.go server/server.go
	@CGO_ENABLED=0 go build
