package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/awesome-gocui/gocui"
)

var demos = map[string]func(){
	"active":        mainActive,
	"bufs":          mainBufs,
	"colors":        mainColors,
	"colors-256":    mainColors256,
	"colors-true":   mainColorsTrue,
	"custom-frames": mainCustomFrames,
	"demo":          mainDemo,
	"dynamic":       mainDynamic,
	"flow-layout":   mainFlowLayout,
	"goroutine":     mainGoroutine,
	"hello":         mainHello,
	"keybinds":      mainKeybinds,
	"layout":        mainLayout,
	"mask":          mainMask,
	"mouse":         mainMouse,
	"on-top":        mainOntop,
	"overlap":       mainOverlap,
	"size":          mainSize,
	"stdin":         mainStdin,
	"table":         mainTable,
	"title":         mainTitle,
	"widgets":       mainWidgets,
	"wrap":          mainWrap,
}

func demoNames() []string {
	names := make([]string, 0)
	for n := range demos {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func usage() string {
	sb := strings.Builder{}
	sb.WriteString("usage:\n")
	sb.WriteString("go run .        : select a demo from the gui and run it\n")
	sb.WriteString("go run . <demo> : run the 'demo' argument\n")
	sb.WriteString("\n")
	sb.WriteString("  where 'demo' can be one of: \n")

	names := demoNames()
	for _, n := range names {
		sb.WriteString("    " + n + "\n")
	}
	return sb.String()
}

func main() {
	demo := ""
	if len(os.Args) > 1 {
		demo = os.Args[1]
	}
	if demo == "-h" || demo == "--help" {
		fmt.Println(usage())
		os.Exit(1)
	}
	if len(demo) > 0 {
		demoFn := demos[demo]
		if demoFn == nil {
			fmt.Println("unknown demo...")
			fmt.Println(usage())
			os.Exit(1)
		}
		demoFn()
		return
	}
	doDemo()
}

type runDemos struct {
	g        *gocui.Gui
	demos    []string
	selected int
}

func doDemo() {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Cursor = false
	g.SelFgColor = gocui.ColorGreen

	d := &runDemos{
		g:     g,
		demos: demoNames(),
	}
	g.SetManagerFunc(d.layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone,
		func(*gocui.Gui, *gocui.View) error {
			return gocui.ErrQuit
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("demos", gocui.KeyArrowDown, gocui.ModNone, d.selectNext); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("demos", gocui.KeyArrowUp, gocui.ModNone, d.selectPrev); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("demos", gocui.KeyEnter, gocui.ModNone, d.runSelectedDemo); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}

}

func (d *runDemos) moveSelected(v *gocui.View, delta int) {
	_ = v.SetHighlight(d.selected, false)
	d.selected += delta
	if d.selected >= len(d.demos) {
		d.selected = 0
	}
	if d.selected < 0 {
		d.selected = len(d.demos) - 1
	}

	x0, y0 := v.Origin()
	_, _, _, y1 := v.Dimensions()
	lcount := y1 - 1

	for d.selected < y0+lcount {
		y0--
		_ = v.SetOrigin(x0, y0)
		_ = v.SetCursor(x0, y0)
	}
	for d.selected >= y0+lcount {
		y0++
		_ = v.SetOrigin(x0, y0)
		_ = v.SetCursor(x0, y0)
	}
	_ = v.SetHighlight(d.selected, true)
}

func (d *runDemos) selectNext(_ *gocui.Gui, v *gocui.View) error {
	d.moveSelected(v, +1)
	return nil
}

func (d *runDemos) selectPrev(_ *gocui.Gui, v *gocui.View) error {
	d.moveSelected(v, -1)
	return nil
}

func (d *runDemos) adjustDim(max, minDim, pos, dim int) (p0 int, p1 int) {
	p0 = pos
	p1 = p0 + dim
	for p1 > max {
		for p0 > 0 {
			p0--
			p1 = p0 + dim
			if p1 < max {
				return
			}
		}
		for dim > minDim {
			dim--
			p1 = p0 + dim
			if p1 < max {
				return
			}
		}
		p1 = max
	}
	return
}

func (d *runDemos) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	x0, x1 := d.adjustDim(maxX, 20, 5, 80)
	y0 := 0
	y1 := y0 + 4

	view, err := g.SetView("help", x0, y0, x1, y1, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		view.Wrap = true
		_, _ = fmt.Fprintln(view, "Use the arrow keys to select a demo")
		_, _ = fmt.Fprintln(view, "Press [Enter] to run the selected demo (Ctrl-C to exit the demo)")
		_, _ = fmt.Fprintln(view, "Ctrl-C to exit")
	}

	y0, y1 = d.adjustDim(maxY, 5, y1+1, len(d.demos)+1)
	view, err = g.SetView("demos", x0, y0, x1, y1, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		//view.Highlight = true
		view.SelBgColor = gocui.ColorGreen
		view.SelFgColor = gocui.ColorBlack

		for i, demo := range d.demos {
			_ = view.SetWritePos(0, i)
			view.WriteString(demo)
		}
		_ = view.SetHighlight(d.selected, true)
		_, _ = g.SetCurrentView("demos")
	}
	return nil
}

func (d *runDemos) runSelectedDemo(_ *gocui.Gui, _ *gocui.View) error {
	name := d.demos[d.selected]

	if name == "stdin" {
		lines := []string{
			"This example doesn't work when running `go run . stdin`",
			"you are supposed to pipe something to this like: `/bin/ls | go run . stdin`",
			"Press 'Esc' to close this view",
		}
		d.messageBox(d.g, 20, 20, "warning", "stdin", lines, "demos")
		return nil
	}
	demoFn := demos[name]

	gocui.Suspend()
	demoFn()
	return gocui.Resume()
}

func (d *runDemos) messageBox(g *gocui.Gui, x0, y0 int, title, viewName string, lines []string, nextView string) {
	w := 20
	for _, l := range lines {
		if len(l) > w {
			w = len(l)
		}
	}
	maxX, maxY := g.Size()
	x0, x1 := d.adjustDim(maxX, 20, x0, w+2)
	y0, y1 := d.adjustDim(maxY, 5, y0, len(lines)+1)

	view, err := g.SetView(viewName, x0, y0, x1, y1, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return
		}

		view.Title = title
		view.Wrap = true
		for _, line := range lines {
			_, _ = fmt.Fprintln(view, line)
		}
		_, _ = g.SetViewOnTop(viewName)
		_, _ = g.SetCurrentView(viewName)
		_ = g.SetKeybinding(viewName, gocui.KeyEsc, gocui.ModNone, func(*gocui.Gui, *gocui.View) error {
			view.Visible = false
			_ = g.DeleteView(viewName)
			_, _ = g.SetCurrentView(nextView)
			return nil
		})
	}
}
