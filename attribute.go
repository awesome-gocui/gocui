// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import (
	"github.com/gdamore/tcell"
)

// Color attributes.
const (
	ColorDefault tcell.Color = tcell.ColorDefault
	ColorBlack               = tcell.ColorBlack
	ColorRed                 = tcell.ColorRed
	ColorGreen               = tcell.ColorGreen
	ColorYellow              = tcell.ColorYellow
	ColorBlue                = tcell.ColorBlue
	ColorMagenta             = tcell.ColorDarkMagenta
	ColorCyan                = tcell.ColorDarkCyan
	ColorWhite               = tcell.ColorWhite
)

// Text style attributes.
const (
	AttrBold      tcell.AttrMask = tcell.AttrBold
	AttrUnderline                = tcell.AttrUnderline
	AttrReverse                  = tcell.AttrReverse
)
