package gocui

type MouseHandler interface {
	Handle(v *View, key Key, ch rune, mod Modifier)
}
