.PHONY: build
build: bin/lanchat

.DEFAULT_GOAL: build

bin/lanchat: main.go ui/*.go lan/*.go logger/*.go
	@CGO_ENABLED=1 go build -race -o ./bin/lanchat ./main.go

.PHONY: test
test:
	go test -race ./...
