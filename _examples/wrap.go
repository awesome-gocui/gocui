// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type demoWrap struct{}

func (d *demoWrap) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("main", 1, 1, maxX-1, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Wrap = true

		line := strings.Repeat("This is a long line -- ", 10)
		_, _ = fmt.Fprintf(v, "%s\n\n", line)
		_, _ = fmt.Fprintln(v, "Short")

		if _, err := g.SetCurrentView("main"); err != nil {
			return err
		}
	}
	return nil
}

func (d *demoWrap) quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func mainWrap() {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	d := &demoWrap{}
	g.SetManagerFunc(d.layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, d.quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}
