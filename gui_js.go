// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build js && wasm
// +build js,wasm

package gocui

import "syscall/js"

// webSetSizeFunc is the js.Func registered as the global tcellSetSize.
// It is kept so a subsequent NewGui can release the previous one.
var webSetSizeFunc js.Func

// getTermWindowSize is get terminal window size on js/wasm.
// There is no tty ioctl available, so ask the tcell screen instead.
//
// If the hosting page defines a terminalCells() function returning
// [cols, rows], the web terminal is resized to that size first, so the
// terminal can fill the browser window instead of tcell's default 80x24.
// The page can also call tcellSetSize(cols, rows) later (e.g. from a
// window resize listener) to resize the running terminal; tcell posts an
// EventResize and gocui relayouts as usual.
func (g *Gui) getTermWindowSize() (int, int, error) {
	applyWebTerminalCells()
	if webSetSizeFunc.Truthy() {
		webSetSizeFunc.Release()
	}
	webSetSizeFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 2 &&
			args[0].Type() == js.TypeNumber && args[1].Type() == js.TypeNumber {
			setScreenSize(args[0].Int(), args[1].Int())
		}
		return nil
	})
	js.Global().Set("tcellSetSize", webSetSizeFunc)
	x, y := screen.Size()
	return x, y, nil
}

// applyWebTerminalCells asks the hosting page for the desired terminal size
// via terminalCells() and applies it. The page is an untrusted boundary:
// anything thrown or malformed (non-array, short array, non-numeric
// elements) is ignored, keeping the current screen size.
func applyWebTerminalCells() {
	defer func() { _ = recover() }()
	fn := js.Global().Get("terminalCells")
	if fn.Type() != js.TypeFunction {
		return
	}
	v := fn.Invoke()
	if v.Type() != js.TypeObject || v.Length() < 2 {
		return
	}
	w, h := v.Index(0), v.Index(1)
	if w.Type() != js.TypeNumber || h.Type() != js.TypeNumber {
		return
	}
	setScreenSize(w.Int(), h.Int())
}

// setScreenSize resizes the tcell screen if the tcell version in use
// supports it (Screen.SetSize was added after the v2.4 pinned here, and
// the WebAssembly backend implements it).
func setScreenSize(w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	if s, ok := screen.(interface{ SetSize(int, int) }); ok {
		s.SetSize(w, h)
	}
}
