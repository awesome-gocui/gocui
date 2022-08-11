// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// WARNING: tricky code just for testing purposes, do not use as reference.

package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/awesome-gocui/gocui"
)

type demoBufs struct {
	vbuf, buf string
}

func (d *demoBufs) quit(_ *gocui.Gui, v *gocui.View) error {
	d.vbuf = v.ViewBuffer()
	d.buf = v.Buffer()
	return gocui.ErrQuit
}

func (d *demoBufs) overwrite(g *gocui.Gui, v *gocui.View) error {
	v.Overwrite = !v.Overwrite
	return nil
}

func (d *demoBufs) layout(g *gocui.Gui) error {
	_, maxY := g.Size()
	if v, err := g.SetView("main", 0, 0, 20, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Editable = true
		v.Wrap = true
		if _, err := g.SetCurrentView("main"); err != nil {
			return err
		}
	}
	return nil
}

func mainBufs() {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}

	d := &demoBufs{}

	g.Cursor = true
	g.Mouse = true

	g.SetManagerFunc(d.layout)

	if err := g.SetKeybinding("main", gocui.KeyCtrlC, gocui.ModNone, d.quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", gocui.KeyCtrlI, gocui.ModNone, d.overwrite); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}

	g.Close()

	fmt.Printf("VBUF:\n%s\n", d.vbuf)
	fmt.Printf("BUF:\n%s\n", d.buf)
}
