// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import (
	"github.com/gdamore/tcell"
)

// Attribute represents a terminal attribute, like color, font style, etc. They
// can be combined using bitwise OR (|). Note that it is not possible to
// combine multiple color attributes.
type Attribute tcell.Color

// Color attributes.
const (
	ColorDefault Attribute = Attribute(tcell.ColorDefault)
	ColorBlack             = Attribute(tcell.ColorBlack)
	ColorRed               = Attribute(tcell.ColorRed)
	ColorGreen             = Attribute(tcell.ColorGreen)
	ColorYellow            = Attribute(tcell.ColorYellow)
	ColorBlue              = Attribute(tcell.ColorBlue)
	ColorMagenta           = Attribute(tcell.ColorDarkMagenta)
	ColorCyan              = Attribute(tcell.ColorDarkCyan)
	ColorWhite             = Attribute(tcell.ColorWhite)
)

// Text style attributes.
const (
	AttrBold      Attribute = Attribute(tcell.AttrBold)
	AttrUnderline           = Attribute(tcell.AttrUnderline)
	AttrReverse             = Attribute(tcell.AttrReverse)
)
