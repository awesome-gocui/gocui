// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"log"

	"github.com/awesome-gocui/gocui"
)

type demoLayout struct{}

func (d *demoLayout) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if _, err := g.SetView("side", -1, -1, int(0.2*float32(maxX)), maxY-5, 0); err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}
	if _, err := g.SetView("main", int(0.2*float32(maxX)), -1, maxX, maxY-5, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		g.SetCurrentView("main")
	}
	if _, err := g.SetView("cmdline", -1, maxY-5, maxX, maxY, 0); err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}

	return nil
}

func (d *demoLayout) quit(*gocui.Gui, *gocui.View) error {
	return gocui.ErrQuit
}

func mainLayout() {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	d := &demoLayout{}
	g.SetManagerFunc(d.layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, d.quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}
