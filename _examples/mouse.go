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

type demoMouse struct {
	initialMouseX, initialMouseY, xOffset, yOffset int
	globalMouseDown, msgMouseDown, movingMsg       bool
}

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
	maxX, maxY := g.Size()
	if _, err := g.View("msg"); d.msgMouseDown && err == nil {
		d.moveMsg(g)
	}
	if v, err := g.SetView("global", -1, -1, maxX, maxY, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = false
	}
	if v, err := g.SetView("but1", 2, 2, 22, 7, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		fmt.Fprintln(v, "Button 1 - line 1")
		fmt.Fprintln(v, "Button 1 - line 2")
		fmt.Fprintln(v, "Button 1 - line 3")
		fmt.Fprintln(v, "Button 1 - line 4")
		if _, err := g.SetCurrentView("but1"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("but2", 24, 2, 44, 4, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		fmt.Fprintln(v, "Button 2 - line 1")
	}
	d.updateHighlightedView(g)
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
	if err := g.SetKeybinding("", gocui.MouseRelease, gocui.ModNone, d.mouseUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, d.globalDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("msg", gocui.MouseLeft, gocui.ModNone, d.msgDown); err != nil {
		return err
	}
	return nil
}

func (d *demoMouse) quit(*gocui.Gui, *gocui.View) error {
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
	if v, err := g.SetView("msg", maxX/2-10, maxY/2, maxX/2+10, maxY/2+2, 0); err == nil || errors.Is(err, gocui.ErrUnknownView) {
		v.Clear()
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack
		fmt.Fprintln(v, l)
	}
	return nil
}

func (d *demoMouse) updateHighlightedView(g *gocui.Gui) {
	mx, my := g.MousePosition()
	for _, view := range g.Views() {
		view.Highlight = false
	}
	if v, err := g.ViewByPosition(mx, my); err == nil {
		v.Highlight = true
	}
}

func (d *demoMouse) moveMsg(g *gocui.Gui) {
	mx, my := g.MousePosition()
	if !d.movingMsg && (mx != d.initialMouseX || my != d.initialMouseY) {
		d.movingMsg = true
	}
	g.SetView("msg", mx-d.xOffset, my-d.yOffset, mx-d.xOffset+20, my-d.yOffset+2, 0)
}

func (d *demoMouse) msgDown(g *gocui.Gui, v *gocui.View) error {
	d.initialMouseX, d.initialMouseY = g.MousePosition()
	if vx, vy, _, _, err := g.ViewPosition("msg"); err == nil {
		d.xOffset = d.initialMouseX - vx
		d.yOffset = d.initialMouseY - vy
		d.msgMouseDown = true
	}
	return nil
}

func (d *demoMouse) globalDown(g *gocui.Gui, v *gocui.View) error {
	mx, my := g.MousePosition()
	if vx0, vy0, vx1, vy1, err := g.ViewPosition("msg"); err == nil {
		if mx >= vx0 && mx <= vx1 && my >= vy0 && my <= vy1 {
			return d.msgDown(g, v)
		}
	}
	d.globalMouseDown = true
	maxX, _ := g.Size()
	msg := fmt.Sprintf("Mouse down at: %d,%d", mx, my)
	x := mx - len(msg)/2
	if x < 0 {
		x = 0
	} else if x+len(msg)+1 > maxX-1 {
		x = maxX - 1 - len(msg) - 1
	}
	if v, err := g.SetView("globalDown", x, my-1, x+len(msg)+1, my+1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.WriteString(msg)
	}
	return nil
}

func (d *demoMouse) mouseUp(g *gocui.Gui, v *gocui.View) error {
	if d.msgMouseDown {
		d.msgMouseDown = false
		if d.movingMsg {
			d.movingMsg = false
			return nil
		} else {
			_ = g.DeleteView("msg")
		}
	} else if d.globalMouseDown {
		d.globalMouseDown = false
		_ = g.DeleteView("globalDown")
	}
	return nil
}
