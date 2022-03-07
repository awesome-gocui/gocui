// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/awesome-gocui/gocui"
)

type demoSize struct{}

func mainSize() {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	d := &demoSize{}
	g.SetManagerFunc(d.layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, d.quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}

func (d *demoSize) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	v, err := g.SetView("size", maxX/2-7, maxY/2, maxX/2+7, maxY/2+2, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		if _, err := g.SetCurrentView("size"); err != nil {
			return err
		}
	}
	v.Clear()
	_, _ = fmt.Fprintf(v, "%d, %d", maxX, maxY)
	return nil
}

func (d *demoSize) quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}
