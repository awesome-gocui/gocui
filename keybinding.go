// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import "github.com/gdamore/tcell"

// Keybidings are used to link a given key-press event with a handler.
type keybinding struct {
	viewName string
	key      tcell.Key
	ch       rune
	mod      tcell.ModMask
	handler  func(*Gui, *View) error
}

// newKeybinding returns a new Keybinding object.
func newKeybinding(viewname string, key tcell.Key, ch rune, mod tcell.ModMask, handler func(*Gui, *View) error) (kb *keybinding) {
	kb = &keybinding{
		viewName: viewname,
		key:      key,
		ch:       ch,
		mod:      mod,
		handler:  handler,
	}
	return kb
}

// matchKeypress returns if the keybinding matches the keypress.
func (kb *keybinding) matchKeypress(key tcell.Key, ch rune, mod tcell.ModMask) bool {
	return kb.key == key && kb.ch == ch && kb.mod == mod
}

// matchView returns if the keybinding matches the current view.
func (kb *keybinding) matchView(v *View) bool {
	// if the user is typing in a field, ignore char keys
	if v == nil || (v.Editable && kb.ch != 0) {
		return false
	}
	return kb.viewName == v.name
}
