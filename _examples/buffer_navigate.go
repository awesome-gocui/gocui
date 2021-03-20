// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/awesome-gocui/gocui"
)

type randomValue struct {
	id    int
	val   string
	prop1 string
	prop2 string
	prop3 string
}

var rndVals = []randomValue{
	{id: 0, val: "first", prop1: "prop first", prop2: "second property", prop3: "next very important data"},
	{id: 1, val: "next", prop1: "test1", prop2: "test2", prop3: "test3"},
	{id: 2, val: "another", prop1: "qqq", prop2: "www", prop3: "eee"},
	{id: 3, val: "very long value", prop1: "very", prop2: "long", prop3: "value"},
	{id: 4, val: "last", prop1: "the", prop2: "last", prop3: "value"},
	{id: 5, val: "first", prop1: "prop first", prop2: "second property", prop3: "next very important data"},
	{id: 6, val: "next", prop1: "test1", prop2: "test2", prop3: "test3"},
	{id: 7, val: "another", prop1: "qqq", prop2: "www", prop3: "eee"},
	{id: 8, val: "very long value", prop1: "very", prop2: "long", prop3: "value"},
	{id: 9, val: "last", prop1: "the", prop2: "last", prop3: "value"},
	{id: 10, val: "first", prop1: "prop first", prop2: "second property", prop3: "next very important data"},
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}

	return g.SetViewOnTop(name)
}

func layout(g *gocui.Gui) error {
	var (
		v1  *gocui.View
		err error
	)

	if v1, err = g.SetView("v1", 0, 0, 20, 12, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		v1.Highlight = true
		v1.Title = "val"
		v1.Wrap = true

		if _, err = setCurrentViewOnTop(g, "v1"); err != nil {
			return err
		}
		str := ""
		for k, r := range rndVals {
			str += fmt.Sprintf("[%d] ID:%d %s\n", k, r.id, r.val)
		}

		fmt.Fprint(v1, strings.TrimSpace(str))
	}

	if v, err := g.SetView("v2", 21, 0, 60, 5, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		v.Title = "props"
		v.Wrap = true

		fmt.Fprint(v, BuffViewV2(&rndVals[0]))
	}

	if v, err := g.SetView("v3", 21, 6, 60, 12, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		v.Wrap = true

		fmt.Fprint(v, BuffViewV3(v1))
	}

	return err
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func cursorDownV1(g *gocui.Gui, v *gocui.View) error {
	v.SetViewLineDown()

	navigateV1(g)

	return nil
}

func cursorUpV1(g *gocui.Gui, v *gocui.View) error {
	v.SetViewLineUp()

	navigateV1(g)

	return nil
}

func mouseSelectV1(g *gocui.Gui, v *gocui.View) error {
	err := mouseSelect(g, v)

	if err != nil {
		return err
	}
	navigateV1(g)

	return err
}

func mouseSelect(g *gocui.Gui, v *gocui.View) (err error) {
	if _, err := g.SetCurrentView(v.Name()); err != nil {
		return err
	}

	return err
}

func navigateV1(g *gocui.Gui) {
	v, _ := g.View("v1")
	pos := v.BufferLinePosition()
	if pos >= 0 {
		fillChildView("v2", g, BuffViewV2(&rndVals[pos]))
	} else {
		fillChildView("v2", g, "")
	}

	fillChildView("v3", g, BuffViewV3(v))
}

func fillChildView(viewName string, g *gocui.Gui, data string) {
	vv, err := g.View(viewName)

	if err != nil {
		return
	}

	if vv != nil {
		vv.Clear()

		if data == "" {
			return
		}
		fmt.Fprint(vv, data)
	}
}

func BuffViewV2(itm *randomValue) (str string) {
	str += fmt.Sprintf("ID: %d\n", itm.id)
	str += fmt.Sprintf("prop1: %s\n", itm.prop1)
	str += fmt.Sprintf("prop2: %s\n", itm.prop2)
	str += fmt.Sprintf("prop3: %s", itm.prop3)

	return str
}

func BuffViewV3(v *gocui.View) (str string) {
	str += fmt.Sprintf("buf line index: %d\n", v.BufferLinePosition())
	str += fmt.Sprintf("buf lines count: %d\n", len(v.BufferLines()))
	str += fmt.Sprintf("view lines count: %d\n", len(v.ViewBufferLines()))
	_, y := v.Cursor()
	str += fmt.Sprintf("cursor y: %d\n", y)
	_, y = v.Origin()
	str += fmt.Sprintf("origin y: %d", y)

	return str
}

func serviceFunc(g *gocui.Gui) {
	time.Sleep(50 * time.Millisecond)

	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("v1")
		if err != nil {
			return nil
		}

		fillChildView("v3", g, BuffViewV3(v))

		return nil
	})
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Cursor = false
	g.Mouse = true
	//g.SelBgColor = gocui.ColorGreen
	g.SelFgColor = gocui.ColorGreen
	g.SelFrameColor = gocui.ColorGreen

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("v1", gocui.KeyArrowDown, gocui.ModNone, cursorDownV1); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("v1", gocui.KeyArrowUp, gocui.ModNone, cursorUpV1); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("v1", gocui.MouseLeft, gocui.ModNone, mouseSelectV1); err != nil {
		log.Panicln(err)
	}

	go serviceFunc(g)

	if err := g.MainLoop(); err != nil && !gocui.IsQuit(err) {
		log.Panicln(err)
	}
}
