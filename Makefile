.PHONY: build

build: lanchat

lanchat: main.go peer.go
	@CGO_ENABLED=0 go build
