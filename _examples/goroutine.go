// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/awesome-gocui/gocui"
)

type demoGoRoutine struct {
	numGoroutines int
	done          chan struct{}
	wg            sync.WaitGroup
	mu            sync.Mutex // protects ctr
	ctr           int
}

func mainGoroutine() {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	d := &demoGoRoutine{
		numGoroutines: 20,
		done:          make(chan struct{}),
		ctr:           0,
	}

	g.SetManagerFunc(d.layout)

	if err := d.keybindings(g); err != nil {
		log.Panicln(err)
	}

	for i := 0; i < d.numGoroutines; i++ {
		d.wg.Add(1)
		go d.counter(g)
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}

	d.wg.Wait()
}

func (d *demoGoRoutine) layout(g *gocui.Gui) error {
	if v, err := g.SetView("ctr", 2, 2, 22, 2+d.numGoroutines+1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Clear()
		if _, err := g.SetCurrentView("ctr"); err != nil {
			return err
		}
	}
	return nil
}

func (d *demoGoRoutine) keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, d.quit); err != nil {
		return err
	}
	return nil
}

func (d *demoGoRoutine) quit(g *gocui.Gui, v *gocui.View) error {
	_ = g
	_ = v
	close(d.done)
	return gocui.ErrQuit
}

func (d *demoGoRoutine) counter(g *gocui.Gui) {
	defer d.wg.Done()

	for {
		select {
		case <-d.done:
			return
		case <-time.After(500 * time.Millisecond):
			d.mu.Lock()
			n := d.ctr
			d.ctr++
			d.mu.Unlock()

			g.Update(func(g *gocui.Gui) error {
				v, err := g.View("ctr")
				if err != nil {
					return err
				}
				// use ctr to make it more chaotic
				// "pseudo-randomly" print in one of two columns (x = 0, and x = 10)
				x := (d.ctr / d.numGoroutines) & 1
				if x != 0 {
					x = 10
				}
				y := d.ctr % d.numGoroutines
				_ = v.SetWritePos(x, y)
				_, _ = fmt.Fprintln(v, n)
				return nil
			})
		}
	}
}
