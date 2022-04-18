package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type PacketType int

const (
	PacketTypeChat = iota
	PacketTypeAdmin
)

const selfColor = "#00ff00"

type Packet struct {
	User string
	Msg  string
	Type PacketType
}

type UI struct {
	FromClient chan Packet
	ToClient   chan Packet
	app        *tview.Application
	chat       *tview.TextView
	input      *tview.InputField
}

func New(user string, fromClient chan Packet, toClient chan Packet) UI {
	grid := tview.NewGrid().SetRows(0, 1)
	chat := newTextView("").Clear()
	app := tview.NewApplication()
	input := newInputField(app, user)
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			msg := input.GetText()
			if msg == "" {
				return
			}
			pkt := Packet{Msg: msg}
			toClient <- pkt
			input.SetText("")
			fmt.Fprintf(chat, "[%s::b]%s> [-:-:-]%s[-:-:-]\n", selfColor, user, pkt.Msg)
			// input.Clear()
		}

	})
	grid.AddItem(chat, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(input, 1, 0, 1, 1, 0, 0, true)
	app.SetRoot(grid, true).SetFocus(input)

	u := UI{FromClient: fromClient, ToClient: toClient, app: app, chat: chat, input: input}
	return u
}

func (u *UI) Run() {
	go u.processPackets()
	err := u.app.Run()

	if err != nil {
		panic(err)
	} else {
		return
	}
}

func newTextView(text string) *tview.TextView {
	return tview.NewTextView().
		SetText("").
		SetDynamicColors(true).
		SetWordWrap(true)
}

func newInputField(app *tview.Application, user string) *tview.InputField {
	return tview.NewInputField().
		SetLabel(fmt.Sprintf("[%s::b]%s> [-:-:-]", selfColor, user))
}

func (u *UI) processPackets() {
	for pkt := range u.FromClient {
		var f func()
		if pkt.Type == PacketTypeChat {
			f = u.drawMsg(pkt)
		} else if pkt.Type == PacketTypeAdmin {
			f = u.drawAdmin(pkt)
		} else {
			// no-op
			f = func() {}
		}
		u.app.QueueUpdateDraw(f)
	}
}

func (u *UI) drawMsg(pkt Packet) func() {
	return func() {
		fmt.Fprintf(u.chat, "[yellow::b]%s> [-:-:-]%s[-:-:-]\n", pkt.User, pkt.Msg)
	}
}

func (u *UI) drawAdmin(pkt Packet) func() {
	return func() {
		fmt.Fprintf(u.chat, "[blue::]-- %s[-:-:-]\n", pkt.Msg)
	}
}
