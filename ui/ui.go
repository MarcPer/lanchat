package ui

import (
	"fmt"
	"io"
	"strings"
)

var prompt string
var promptDelete string
var outWriter io.Writer

const (
	fontReset  = "\033[0m"
	fontBold   = "\033[1m"
	fontGreen  = "\033[92m"
	fontBlue   = "\033[34m"
	fontYellow = "\033[93m"
)

type Packet struct {
	User string
	Msg  string
}

func SetUserPrompt(username string) {
	prompt = fmt.Sprintf("%s%s%s>%s ", fontBold, fontGreen, username, fontReset)
	promptDelete = strings.Repeat("\b", len(prompt))
}

func printAdmin(msg string) {
	fmt.Fprint(outWriter, promptDelete, strings.Repeat(" ", len(promptDelete)), promptDelete)
	fmt.Fprintln(outWriter, msg)
	fmt.Fprint(outWriter, prompt)
}

func printUser(self bool, username, msg string) {
	fmt.Fprint(outWriter, promptDelete, strings.Repeat(" ", len(promptDelete)), promptDelete)

	fmt.Print(prompt)
}
