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

type demoMouse struct{}

func mainMouse() {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = false
	g.Mouse = true

	d := demoMouse{}
	g.SetManagerFunc(d.layout)

	if err := d.keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}

func (d *demoMouse) layout(g *gocui.Gui) error {
	if v, err := g.SetView("but1", 2, 2, 22, 7, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		_, _ = fmt.Fprintln(v, "Button 1 - line 1")
		_, _ = fmt.Fprintln(v, "Button 1 - line 2")
		_, _ = fmt.Fprintln(v, "Button 1 - line 3")
		_, _ = fmt.Fprintln(v, "Button 1 - line 4")
		if _, err := g.SetCurrentView("but1"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("but2", 24, 2, 44, 4, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		_, _ = fmt.Fprintln(v, "Button 2 - line 1")
	}
	return nil
}

func (d *demoMouse) keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, d.quit); err != nil {
		return err
	}
	for _, n := range []string{"but1", "but2"} {
		if err := g.SetKeybinding(n, gocui.MouseLeft, gocui.ModNone, d.showMsg); err != nil {
			return err
		}
	}
	if err := g.SetKeybinding("msg", gocui.MouseLeft, gocui.ModNone, d.delMsg); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.MouseRight, gocui.ModNone, d.delMsg); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.MouseMiddle, gocui.ModNone, d.delMsg); err != nil {
		return err
	}
	return nil
}

func (d *demoMouse) quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func (d *demoMouse) showMsg(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	if _, err := g.SetCurrentView(v.Name()); err != nil {
		return err
	}

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	maxX, maxY := g.Size()
	if v, err := g.SetView("msg", maxX/2-10, maxY/2, maxX/2+10, maxY/2+2, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		_, _ = fmt.Fprintln(v, l)
	}
	return nil
}

func (d *demoMouse) delMsg(g *gocui.Gui, _ *gocui.View) error {
	// Error check removed, because delete could be called multiple times with the above keybindings
	_ = g.DeleteView("msg")
	return nil
}
