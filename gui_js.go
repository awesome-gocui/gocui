// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build js
// +build js

package gocui

// getTermWindowSize is get terminal window size on js/wasm.
// There is no tty ioctl available, so ask the tcell screen instead.
func (g *Gui) getTermWindowSize() (int, int, error) {
	x, y := screen.Size()
	return x, y, nil
}
