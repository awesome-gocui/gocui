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

type demoDynamic struct {
	views   []string
	curView int
	idxView int
	delta   int
}

func mainDynamic() {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.SelFgColor = gocui.ColorRed
	g.SelFrameColor = gocui.ColorRed

	d := &demoDynamic{
		views:   []string{},
		curView: -1,
		idxView: 0,
		delta:   1,
	}
	g.SetManagerFunc(d.layout)

	if err := d.initKeybindings(g); err != nil {
		log.Panicln(err)
	}
	if err := d.newView(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}

func (d *demoDynamic) layout(g *gocui.Gui) error {
	maxX, _ := g.Size()
	v, err := g.SetView("help", maxX-25, 0, maxX-1, 9, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		fmt.Fprintln(v, "KEYBINDINGS")
		fmt.Fprintln(v, "Space: New View")
		fmt.Fprintln(v, "Tab: Next View")
		fmt.Fprintln(v, "← ↑ → ↓: Move View")
		fmt.Fprintln(v, "Backspace: Delete View")
		fmt.Fprintln(v, "t: Set view on top")
		fmt.Fprintln(v, "b: Set view on bottom")
		fmt.Fprintln(v, "^C: Exit")
	}
	return nil
}

func (d *demoDynamic) initKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return gocui.ErrQuit
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeySpace, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return d.newView(g)
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyBackspace, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return d.delView(g)
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyBackspace2, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return d.delView(g)
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return d.nextView(g, true)
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return d.moveView(g, v, -d.delta, 0)
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return d.moveView(g, v, d.delta, 0)
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return d.moveView(g, v, 0, d.delta)
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return d.moveView(g, v, 0, -d.delta)
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 't', gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			_, err := g.SetViewOnTop(d.views[d.curView])
			return err
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'b', gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			_, err := g.SetViewOnBottom(d.views[d.curView])
			return err
		}); err != nil {
		return err
	}
	return nil
}

func (d *demoDynamic) newView(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	name := fmt.Sprintf("v%v", d.idxView)
	v, err := g.SetView(name, maxX/2-5, maxY/2-5, maxX/2+5, maxY/2+5, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Wrap = true
		fmt.Fprintln(v, strings.Repeat(name+" ", 30))
	}
	if _, err := g.SetCurrentView(name); err != nil {
		return err
	}

	d.views = append(d.views, name)
	d.curView = len(d.views) - 1
	d.idxView += 1
	return nil
}

func (d *demoDynamic) delView(g *gocui.Gui) error {
	if len(d.views) <= 1 {
		return nil
	}

	if err := g.DeleteView(d.views[d.curView]); err != nil {
		return err
	}
	d.views = append(d.views[:d.curView], d.views[d.curView+1:]...)

	return d.nextView(g, false)
}

func (d *demoDynamic) nextView(g *gocui.Gui, disableCurrent bool) error {
	next := d.curView + 1
	if next > len(d.views)-1 {
		next = 0
	}

	if _, err := g.SetCurrentView(d.views[next]); err != nil {
		return err
	}

	d.curView = next
	return nil
}

func (d *demoDynamic) moveView(g *gocui.Gui, v *gocui.View, dx, dy int) error {
	name := v.Name()
	x0, y0, x1, y1, err := g.ViewPosition(name)
	if err != nil {
		return err
	}
	if _, err := g.SetView(name, x0+dx, y0+dy, x1+dx, y1+dy, 0); err != nil {
		return err
	}
	return nil
}
