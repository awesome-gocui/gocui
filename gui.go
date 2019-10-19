// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import (
	standardErrors "errors"

	"github.com/gdamore/tcell"
	"github.com/go-errors/errors"
)

// OutputMode represents the terminal's output mode (8 or 256 colors).
// TODO: Fix this
// type OutputMode tcell.Color

var (
	// ErrQuit is used to decide if the MainLoop finished successfully.
	ErrQuit = standardErrors.New("quit")

	// ErrUnknownView allows to assert if a View must be initialized.
	ErrUnknownView = standardErrors.New("unknown view")
)

// TODO: Fix this
// const (
// 	// OutputNormal provides 8-colors terminal mode.
// 	OutputNormal = OutputMode(termbox.OutputNormal)

// 	// Output256 provides 256-colors terminal mode.
// 	Output256 = OutputMode(termbox.Output256)

// 	// OutputGrayScale provides greyscale terminal mode.
// 	OutputGrayScale = OutputMode(termbox.OutputGrayscale)

// 	// Output216 provides greyscale terminal mode.
// 	Output216 = OutputMode(termbox.Output216)
// )

// Gui represents the whole User Interface, including the views, layouts
// and keybindings.
type Gui struct {
	rerenderEvents chan struct{}
	userEvents     chan userEvent
	tcellEvent     chan tcell.Event
	views          []*View
	currentView    *View
	managers       []Manager
	keybindings    []*keybinding
	maxX, maxY     int
	screen         tcell.Screen
	stop           chan struct{}

	// BgColor and FgColor allow to configure the background and foreground
	// colors of the GUI.
	BgColor, FgColor, FrameColor tcell.Color

	// SelBgColor and SelFgColor allow to configure the background and
	// foreground colors of the frame of the current view.
	SelBgColor, SelFgColor, SelFrameColor tcell.Color

	// If Highlight is true, Sel{Bg,Fg}Colors will be used to draw the
	// frame of the current view.
	Highlight bool

	// If Cursor is true then the cursor is enabled.
	Cursor bool

	// If Mouse is true then mouse events will be enabled.
	Mouse bool

	// If InputEsc is true, when ESC sequence is in the buffer and it doesn't
	// match any known sequence, ESC means KeyEsc.
	InputEsc bool

	// SupportOverlaps is true when we allow for view edges to overlap with other
	// view edges
	SupportOverlaps bool
}

// NewGui returns a new Gui object with a given output mode.
func NewGui(supportOverlaps bool) (*Gui, error) {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	err = screen.Init()
	if err != nil {
		return nil, err
	}

	// NOTE: We might not need this
	screen.SetStyle(tcell.StyleDefault)

	screen.Clear()

	g := &Gui{screen: screen}

	g.stop = make(chan struct{})

	g.rerenderEvents = make(chan struct{})
	g.userEvents = make(chan userEvent, 20)
	g.tcellEvent = make(chan tcell.Event)

	g.maxX, g.maxY = screen.Size()

	g.BgColor, g.FgColor, g.FrameColor = ColorDefault, ColorDefault, ColorDefault
	g.SelBgColor, g.SelFgColor, g.SelFrameColor = ColorDefault, ColorDefault, ColorDefault

	// SupportOverlaps is true when we allow for view edges to overlap with other
	// view edges
	g.SupportOverlaps = supportOverlaps

	g.RegisterFallbacks()

	return g, nil
}

// RegisterFallbacks adds fallback chars in case only ASCII is supported
func (g *Gui) RegisterFallbacks() {
	fallbacks := map[rune]string{
		'─': "-",
		'│': "|",
		'┌': "+",
		'┐': "+",
		'└': "+",
		'┘': "+",
	}
	for r, fallback := range fallbacks {
		g.screen.RegisterRuneFallback(r, fallback)
	}
}

// Close finalizes the library. It should be called after a successful
// initialization and when gocui is not needed anymore.
func (g *Gui) Close() {
	go func() {
		g.stop <- struct{}{}
	}()
	g.screen.Fini()
}

// Size returns the terminal's size.
func (g *Gui) Size() (x, y int) {
	return g.maxX, g.maxY
}

// SetRune writes a rune at the given point, relative to the top-left
// corner of the terminal. It checks if the position is valid and applies
// the given colors.
func (g *Gui) SetRune(x, y int, ch rune, fgColor, bgColor tcell.Color) error {
	if x < 0 || y < 0 || x >= g.maxX || y >= g.maxY {
		return errors.New("invalid point")
	}

	g.screen.SetCell(x, y, tcell.StyleDefault.Foreground(fgColor).Background(bgColor), ch)
	return nil
}

// Rune returns the rune contained in the cell at the given position.
// It checks if the position is valid.
func (g *Gui) Rune(x, y int) (rune, error) {
	if x < 0 || y < 0 || x >= g.maxX || y >= g.maxY {
		return ' ', errors.New("invalid point")
	}

	char, _, _, _ := g.screen.GetContent(x, y)
	return char, nil
}

// SetView creates a new view with its top-left corner at (x0, y0)
// and the bottom-right one at (x1, y1). If a view with the same name
// already exists, its dimensions are updated; otherwise, the error
// ErrUnknownView is returned, which allows to assert if the View must
// be initialized. It checks if the position is valid.
func (g *Gui) SetView(name string, x0, y0, x1, y1 int, overlaps byte) (*View, error) {
	if x0 >= x1 {
		return nil, errors.New("invalid dimensions")
	}
	if name == "" {
		return nil, errors.New("invalid name")
	}

	if v, err := g.View(name); err == nil {
		v.x0 = x0
		v.y0 = y0
		v.x1 = x1
		v.y1 = y1
		v.tainted = true
		return v, nil
	}

	v := newView(g.screen, name, x0, y0, x1, y1)
	v.BgColor, v.FgColor = g.BgColor, g.FgColor
	v.SelBgColor, v.SelFgColor = g.SelBgColor, g.SelFgColor
	v.Overlaps = overlaps
	g.views = append(g.views, v)
	return v, errors.Wrap(ErrUnknownView, 0)
}

// SetViewOnTop sets the given view on top of the existing ones.
func (g *Gui) SetViewOnTop(name string) (*View, error) {
	for i, v := range g.views {
		if v.name == name {
			s := append(g.views[:i], g.views[i+1:]...)
			g.views = append(s, v)
			return v, nil
		}
	}
	return nil, errors.Wrap(ErrUnknownView, 0)
}

// SetViewOnBottom sets the given view on bottom of the existing ones.
func (g *Gui) SetViewOnBottom(name string) (*View, error) {
	for i, v := range g.views {
		if v.name == name {
			s := append(g.views[:i], g.views[i+1:]...)
			g.views = append([]*View{v}, s...)
			return v, nil
		}
	}
	return nil, errors.Wrap(ErrUnknownView, 0)
}

// Views returns all the views in the GUI.
func (g *Gui) Views() []*View {
	return g.views
}

// View returns a pointer to the view with the given name, or error
// ErrUnknownView if a view with that name does not exist.
func (g *Gui) View(name string) (*View, error) {
	for _, v := range g.views {
		if v.name == name {
			return v, nil
		}
	}
	return nil, errors.Wrap(ErrUnknownView, 0)
}

// ViewByPosition returns a pointer to a view matching the given position, or
// error ErrUnknownView if a view in that position does not exist.
func (g *Gui) ViewByPosition(x, y int) (*View, error) {
	// traverse views in reverse order checking top views first
	for i := len(g.views); i > 0; i-- {
		v := g.views[i-1]
		if x > v.x0 && x < v.x1 && y > v.y0 && y < v.y1 {
			return v, nil
		}
	}
	return nil, errors.Wrap(ErrUnknownView, 0)
}

// ViewPosition returns the coordinates of the view with the given name, or
// error ErrUnknownView if a view with that name does not exist.
func (g *Gui) ViewPosition(name string) (x0, y0, x1, y1 int, err error) {
	for _, v := range g.views {
		if v.name == name {
			return v.x0, v.y0, v.x1, v.y1, nil
		}
	}
	return 0, 0, 0, 0, errors.Wrap(ErrUnknownView, 0)
}

// DeleteView deletes a view by name.
func (g *Gui) DeleteView(name string) error {
	for i, v := range g.views {
		if v.name == name {
			g.views = append(g.views[:i], g.views[i+1:]...)
			return nil
		}
	}
	return errors.Wrap(ErrUnknownView, 0)
}

// SetCurrentView gives the focus to a given view.
func (g *Gui) SetCurrentView(name string) (*View, error) {
	for _, v := range g.views {
		if v.name == name {
			g.currentView = v
			return v, nil
		}
	}
	return nil, errors.Wrap(ErrUnknownView, 0)
}

// CurrentView returns the currently focused view, or nil if no view
// owns the focus.
func (g *Gui) CurrentView() *View {
	return g.currentView
}

// SetKeybinding creates a new keybinding. If viewname equals to ""
// (empty string) then the keybinding will apply to all views. key must
// be a rune or a Key.
func (g *Gui) SetKeybinding(viewname string, key interface{}, mod tcell.ModMask, handler func(*Gui, *View) error) error {
	k, ch, err := getKey(key)
	if err != nil {
		return err
	}
	kb := newKeybinding(viewname, k, ch, mod, handler)
	g.keybindings = append(g.keybindings, kb)
	return nil
}

// DeleteKeybinding deletes a keybinding.
func (g *Gui) DeleteKeybinding(viewname string, key interface{}, mod tcell.ModMask) error {
	k, ch, err := getKey(key)
	if err != nil {
		return err
	}

	for i, kb := range g.keybindings {
		if kb.viewName == viewname && kb.ch == ch && kb.key == k && kb.mod == mod {
			g.keybindings = append(g.keybindings[:i], g.keybindings[i+1:]...)
			return nil
		}
	}
	return errors.New("keybinding not found")
}

// DeleteKeybindings deletes all keybindings of view.
func (g *Gui) DeleteKeybindings(viewname string) {
	var s []*keybinding
	for _, kb := range g.keybindings {
		if kb.viewName != viewname {
			s = append(s, kb)
		}
	}
	g.keybindings = s
}

// getKey takes an empty interface with a key and returns the corresponding
// typed Key or rune.
func getKey(key interface{}) (tcell.Key, rune, error) {
	switch t := key.(type) {
	case tcell.Key:
		return t, 0, nil
	case rune:
		return 0, t, nil
	default:
		return 0, 0, errors.New("unknown type")
	}
}

// userEvent represents an event triggered by the user.
type userEvent struct {
	f func(*Gui) error
}

// Update executes the passed function. This method can be called safely from a
// goroutine in order to update the GUI. It is important to note that the
// passed function won't be executed immediately, instead it will be added to
// the user events queue. Given that Update spawns a goroutine, the order in
// which the user events will be handled is not guaranteed.
func (g *Gui) Update(f func(*Gui) error) {
	go func() { g.userEvents <- userEvent{f: f} }()
}

// A Manager is in charge of GUI's layout and can be used to build widgets.
type Manager interface {
	// Layout is called every time the GUI is redrawn, it must contain the
	// base views and its initializations.
	Layout(*Gui) error
}

// The ManagerFunc type is an adapter to allow the use of ordinary functions as
// Managers. If f is a function with the appropriate signature, ManagerFunc(f)
// is an Manager object that calls f.
type ManagerFunc func(*Gui) error

// Layout calls f(g)
func (f ManagerFunc) Layout(g *Gui) error {
	return f(g)
}

// SetManager sets the given GUI managers. It deletes all views and
// keybindings.
func (g *Gui) SetManager(managers ...Manager) {
	g.managers = managers
	g.currentView = nil
	g.views = nil
	g.keybindings = nil

	go func() {
		g.rerenderEvents <- struct{}{}
	}()
}

// SetManagerFunc sets the given manager function. It deletes all views and
// keybindings.
func (g *Gui) SetManagerFunc(manager func(*Gui) error) {
	g.SetManager(ManagerFunc(manager))
}

// MainLoop runs the main loop until an error is returned. A successful
// finish should return ErrQuit.
func (g *Gui) MainLoop() error {
	g.loaderTick()
	if err := g.flush(); err != nil {
		return err
	}

	go func() {
		for {
			g.tcellEvent <- g.screen.PollEvent()
		}
	}()

	// TODO: Fix this
	// inputMode = termbox.InputEsc
	// if g.Mouse {
	// 	inputMode |= termbox.InputMouse
	// }
	// termbox.SetInputMode(inputMode)

	if err := g.flush(); err != nil {
		return err
	}
	for {
		select {
		case <-g.stop:
			return nil
		case ev := <-g.tcellEvent:
			if err := g.handleEvent(ev); err != nil {
				return err
			}
		case ev := <-g.userEvents:
			if err := ev.f(g); err != nil {
				return err
			}
		}
		if err := g.consumeevents(); err != nil {
			return err
		}
		if err := g.flush(); err != nil {
			return err
		}
	}
}

// consumeevents handles the remaining events in the events pool.
func (g *Gui) consumeevents() error {
	for {
		select {
		case <-g.rerenderEvents:
			if err := g.handleEvent(nil); err != nil {
				return err
			}
		case ev := <-g.userEvents:
			if err := ev.f(g); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

// handleEvent handles an event, based on its type (key-press, error, etc.)
func (g *Gui) handleEvent(ev tcell.Event) error {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		return g.onKey(ev)
	case *tcell.EventMouse:
		return g.onMouse(ev)
	case *tcell.EventError:
		return ev
	case *tcell.EventResize:
		return g.flush()
	default:
		return nil
	}
}

// flush updates the gui, re-drawing frames and buffers.
func (g *Gui) flush() error {
	g.screen.Clear()

	maxX, maxY := g.screen.Size()
	// if GUI's size has changed, we need to redraw all views
	if maxX != g.maxX || maxY != g.maxY {
		for _, v := range g.views {
			v.tainted = true
		}
	}
	g.maxX, g.maxY = maxX, maxY

	for _, m := range g.managers {
		if err := m.Layout(g); err != nil {
			return err
		}
	}
	for _, v := range g.views {
		if !v.Visible || v.y1 < v.y0 {
			continue
		}
		if v.Frame {
			var fgColor, bgColor, frameColor tcell.Color
			if g.Highlight && v == g.currentView {
				fgColor = g.SelFgColor
				bgColor = g.SelBgColor
				frameColor = g.SelFrameColor
			} else {
				fgColor = g.FgColor
				bgColor = g.BgColor
				frameColor = g.FrameColor
			}

			if err := g.drawFrameEdges(v, frameColor, bgColor); err != nil {
				return err
			}
			if err := g.drawFrameCorners(v, frameColor, bgColor); err != nil {
				return err
			}
			if v.Title != "" {
				if err := g.drawTitle(v, fgColor, bgColor); err != nil {
					return err
				}
			}
			if v.Subtitle != "" {
				if err := g.drawSubtitle(v, fgColor, bgColor); err != nil {
					return err
				}
			}
		}
		if err := g.draw(v); err != nil {
			return err
		}
	}
	g.screen.Sync()
	return nil
}

// drawFrameEdges draws the horizontal and vertical edges of a view.
func (g *Gui) drawFrameEdges(v *View, fgColor, bgColor tcell.Color) error {
	runeH, runeV := '─', '│'

	for x := v.x0 + 1; x < v.x1 && x < g.maxX; x++ {
		if x < 0 {
			continue
		}
		if v.y0 > -1 && v.y0 < g.maxY {
			if err := g.SetRune(x, v.y0, runeH, fgColor, bgColor); err != nil {
				return err
			}
		}
		if v.y1 > -1 && v.y1 < g.maxY {
			if err := g.SetRune(x, v.y1, runeH, fgColor, bgColor); err != nil {
				return err
			}
		}
	}
	for y := v.y0 + 1; y < v.y1 && y < g.maxY; y++ {
		if y < 0 {
			continue
		}
		if v.x0 > -1 && v.x0 < g.maxX {
			if err := g.SetRune(v.x0, y, runeV, fgColor, bgColor); err != nil {
				return err
			}
		}
		if v.x1 > -1 && v.x1 < g.maxX {
			if err := g.SetRune(v.x1, y, runeV, fgColor, bgColor); err != nil {
				return err
			}
		}
	}
	return nil
}

func cornerRune(index byte) rune {
	return []rune{' ', '│', '│', '│', '─', '┘', '┐', '┤', '─', '└', '┌', '├', '├', '┴', '┬', '┼'}[index]
}

func corner(v *View, directions byte) rune {
	index := v.Overlaps | directions
	return cornerRune(index)
}

// drawFrameCorners draws the corners of the view.
func (g *Gui) drawFrameCorners(v *View, fgColor, bgColor tcell.Color) error {
	if v.y0 == v.y1 {
		if !g.SupportOverlaps && v.x0 >= 0 && v.x1 >= 0 && v.y0 >= 0 && v.x0 < g.maxX && v.x1 < g.maxX && v.y0 < g.maxY {
			if err := g.SetRune(v.x0, v.y0, '╶', fgColor, bgColor); err != nil {
				return err
			}
			if err := g.SetRune(v.x1, v.y0, '╴', fgColor, bgColor); err != nil {
				return err
			}
		}
		return nil
	}

	runeTL, runeTR, runeBL, runeBR := '┌', '┐', '└', '┘'
	if g.SupportOverlaps {
		runeTL = corner(v, BOTTOM|RIGHT)
		runeTR = corner(v, BOTTOM|LEFT)
		runeBL = corner(v, TOP|RIGHT)
		runeBR = corner(v, TOP|LEFT)
	}

	corners := []struct {
		x, y int
		ch   rune
	}{{v.x0, v.y0, runeTL}, {v.x1, v.y0, runeTR}, {v.x0, v.y1, runeBL}, {v.x1, v.y1, runeBR}}

	for _, c := range corners {
		if c.x >= 0 && c.y >= 0 && c.x < g.maxX && c.y < g.maxY {
			if err := g.SetRune(c.x, c.y, c.ch, fgColor, bgColor); err != nil {
				return err
			}
		}
	}
	return nil
}

// drawTitle draws the title of the view.
func (g *Gui) drawTitle(v *View, fgColor, bgColor tcell.Color) error {
	if v.y0 < 0 || v.y0 >= g.maxY {
		return nil
	}

	for i, ch := range v.Title {
		x := v.x0 + i + 2
		if x < 0 {
			continue
		} else if x > v.x1-2 || x >= g.maxX {
			break
		}
		if err := g.SetRune(x, v.y0, ch, fgColor, bgColor); err != nil {
			return err
		}
	}
	return nil
}

// drawSubtitle draws the subtitle of the view.
func (g *Gui) drawSubtitle(v *View, fgColor, bgColor tcell.Color) error {
	if v.y0 < 0 || v.y0 >= g.maxY {
		return nil
	}

	start := v.x1 - 5 - len(v.Subtitle)
	if start < v.x0 {
		return nil
	}
	for i, ch := range v.Subtitle {
		x := start + i
		if x >= v.x1 {
			break
		}
		if err := g.SetRune(x, v.y0, ch, fgColor, bgColor); err != nil {
			return err
		}
	}
	return nil
}

// draw manages the cursor and calls the draw function of a view.
func (g *Gui) draw(v *View) error {
	if g.Cursor {
		if curview := g.currentView; curview != nil {
			vMaxX, vMaxY := curview.Size()
			if curview.cx < 0 {
				curview.cx = 0
			} else if curview.cx >= vMaxX {
				curview.cx = vMaxX - 1
			}
			if curview.cy < 0 {
				curview.cy = 0
			} else if curview.cy >= vMaxY {
				curview.cy = vMaxY - 1
			}

			gMaxX, gMaxY := g.Size()
			cx, cy := curview.x0+curview.cx+1, curview.y0+curview.cy+1
			if cx >= 0 && cx < gMaxX && cy >= 0 && cy < gMaxY {
				g.screen.ShowCursor(cx, cy)
			} else {
				g.screen.HideCursor()
			}
		}
	} else {
		g.screen.HideCursor()
	}

	v.clearRunes()
	if err := v.draw(); err != nil {
		return err
	}
	return nil
}

func (g *Gui) onMouse(ev *tcell.EventMouse) error {
	mouseX, mouseY := ev.Position()
	v, err := g.ViewByPosition(mouseX, mouseY)
	if err != nil {
		return err
	}
	if err := v.SetCursor(mouseX-v.x0-1, mouseY-v.y0-1); err != nil {
		return err
	}
	_, err = g.execKeybindings(v, ev)
	return err
}

// onKey manages key-press events.
// A keybinding handler is called when a key-press or mouse event satisfies
// a configured keybinding. Furthermore, currentView's internal buffer is
// modified if currentView.Editable is true.
func (g *Gui) onKey(ev *tcell.EventKey) error {
	matched, err := g.execKeybindings(g.currentView, ev)
	if err != nil {
		return err
	}
	if matched {
		return nil
	}
	if g.currentView != nil && g.currentView.Editable && g.currentView.Editor != nil {
		g.currentView.Editor.Edit(g.currentView, ev.Key(), ev.Rune(), ev.Modifiers())
	}
	return nil
}

// execKeybindings executes the keybinding handlers that match the passed view
// and event. The value of matched is true if there is a match and no errors.
func (g *Gui) execKeybindings(v *View, ev tcell.Event) (matched bool, err error) {
	var globalKb *keybinding

	for _, kb := range g.keybindings {
		if kb.handler == nil {
			continue
		}

		switch ev := ev.(type) {
		case *tcell.EventKey:
			if !kb.matchKeypress(ev) {
				continue
			}
		case *tcell.EventMouse:
			// TODO: inplement this
			continue
		}

		if kb.matchView(v) {
			return g.execKeybinding(v, kb)
		}
		if kb.viewName == "" && ((v != nil && !v.Editable) || kb.ch == 0) {
			globalKb = kb
		}
	}

	if globalKb != nil {
		return g.execKeybinding(v, globalKb)
	}
	return false, nil
}

// execKeybinding executes a given keybinding
func (g *Gui) execKeybinding(v *View, kb *keybinding) (bool, error) {
	if err := kb.handler(g, v); err != nil {
		return false, err
	}
	return true, nil
}

// IsUnknownView reports whether the contents of an error is "unknown view".
func IsUnknownView(err error) bool {
	return err != nil && err.Error() == ErrUnknownView.Error()
}

// IsQuit reports whether the contents of an error is "quit".
func IsQuit(err error) bool {
	return err != nil && err.Error() == ErrQuit.Error()
}
