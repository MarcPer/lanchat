package ui

import "fmt"

func (u *UI) Write(b []byte) (int, error) {
	fmt.Fprintf(u.chat, "[-:-:-]%s[-:-:-]\n", string(b))
	u.app.Draw()
	return len(b), nil
}
