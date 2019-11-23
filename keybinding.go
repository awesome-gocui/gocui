// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import (
	"strings"

	"github.com/gdamore/tcell"
)

// Keybidings are used to link a given key-press event with a handler.
type keybinding struct {
	viewName string
	key      tcell.Key
	ch       rune
	mod      tcell.ModMask
	handler  func(*Gui, *View) error
}

// Parse takes the input string and extracts the keybinding.
// Returns a Key / rune, a Modifier and an error.
func Parse(input string) (interface{}, tcell.ModMask, error) {
	if len(input) == 1 {
		_, r, err := getKey(rune(input[0]))
		if err != nil {
			return nil, tcell.ModNone, err
		}
		return r, tcell.ModNone, nil
	}

	var modifier tcell.ModMask
	cleaned := make([]string, 0)

	tokens := strings.Split(input, "+")
	for _, t := range tokens {
		normalized := strings.Title(strings.ToLower(t))
		if t == "Alt" {
			modifier = tcell.ModAlt
			continue
		}
		cleaned = append(cleaned, normalized)
	}

	exist := false
	var key tcell.Key
	for mapKey, name := range tcell.KeyNames {
		if name == strings.Join(cleaned, "-") {
			key = mapKey
			exist = true
			break
		}
	}
	if !exist {
		return nil, tcell.ModNone, ErrNoSuchKeybind
	}

	return key, modifier, nil
}

// ParseAll takes an array of strings and returns a map of all keybindings.
func ParseAll(input []string) (map[interface{}]tcell.ModMask, error) {
	ret := make(map[interface{}]tcell.ModMask)
	for _, i := range input {
		k, m, err := Parse(i)
		if err != nil {
			return ret, err
		}
		ret[k] = m
	}
	return ret, nil
}

// MustParse takes the input string and returns a Key / rune and a Modifier.
// It will panic if any error occured.
func MustParse(input string) (interface{}, tcell.ModMask) {
	k, m, err := Parse(input)
	if err != nil {
		panic(err)
	}
	return k, m
}

// MustParseAll takes an array of strings and returns a map of all keybindings.
// It will panic if any error occured.
func MustParseAll(input []string) map[interface{}]tcell.ModMask {
	result, err := ParseAll(input)
	if err != nil {
		panic(err)
	}
	return result
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
func (kb *keybinding) matchKeypress(keyEvent *tcell.EventKey) bool {
	if kb.mod > 0 && kb.mod != keyEvent.Modifiers() {
		return false
	}
	return (kb.key != 0 && kb.key == keyEvent.Key()) || (kb.ch != 0 && kb.ch == keyEvent.Rune())
}

// matchView returns if the keybinding matches the current view.
func (kb *keybinding) matchView(v *View) bool {
	// if the user is typing in a field, ignore char keys
	if v == nil || (v.Editable && kb.ch != 0) {
		return false
	}
	return kb.viewName == v.name
}
