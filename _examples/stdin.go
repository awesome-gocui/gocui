// Copyright 2015 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This example doesn't work when running `go run stdin.go`, you are suposed to pipe someting to this like: `/bin/ls | go run stdin.go`

package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/awesome-gocui/gocui"
	"github.com/gdamore/tcell"
)

func main() {
	g, err := gocui.NewGui(true)
	if err != nil {
		log.Fatalln(err)
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)

	if err := initKeybindings(g); err != nil {
		log.Fatalln(err)
	}

	if err := g.MainLoop(); err != nil && !gocui.IsQuit(err) {
		log.Fatalln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, _ := g.Size()

	if v, err := g.SetView("help", maxX-23, 0, maxX-1, 5, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		fmt.Fprintln(v, "KEYBINDINGS")
		fmt.Fprintln(v, "↑ ↓: Seek input")
		fmt.Fprintln(v, "a: Enable autoscroll")
		fmt.Fprintln(v, "^C: Exit")
	}

	if v, err := g.SetView("stdin", 0, 0, 80, 35, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		v.Wrap = true

		if _, err := io.Copy(hex.Dumper(v), os.Stdin); err != nil {
			return err
		}

		if _, err := g.SetCurrentView("stdin"); err != nil {
			return err
		}
	}

	return nil
}

func initKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", tcell.KeyCtrlC, tcell.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("stdin", 'a', tcell.ModNone, autoscroll); err != nil {
		return err
	}
	if err := g.SetKeybinding("stdin", tcell.KeyUp, tcell.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			scrollView(v, -1)
			return nil
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("stdin", tcell.KeyDown, tcell.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			scrollView(v, 1)
			return nil
		}); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func autoscroll(g *gocui.Gui, v *gocui.View) error {
	v.Autoscroll = true
	return nil
}

func scrollView(v *gocui.View, dy int) error {
	if v != nil {
		v.Autoscroll = false
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+dy); err != nil {
			return err
		}
	}
	return nil
}
