// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

var (
	// ErrInvalidPoint is returned when client passed invalid coordinates of a cell.
	// Most likely client has passed negative coordinates of a cell.
	ErrInvalidPoint = errors.New("invalid point")
)

// A View is a window. It maintains its own internal buffer and cursor
// position.
type View struct {
	name           string
	x0, y0, x1, y1 int // left top right bottom
	ox, oy         int // view offsets
	cx, cy         int // cursor position
	rx, ry         int // Read() offsets
	wx, wy         int // Write() offsets
	lines          [][]cell

	readBuffer []byte // used for storing unreaded bytes

	tainted   bool       // true if the viewLines must be updated
	viewLines []viewLine // internal representation of the view's buffer
	visible   bool       // marks if the view buffer is visible

	ei *escapeInterpreter // used to decode ESC sequences on Write

	// BgColor and FgColor allow to configure the background and foreground
	// colors of the View.
	BgColor, FgColor Attribute

	// SelBgColor and SelFgColor are used to configure the background and
	// foreground colors of the selected line, when it is highlighted.
	SelBgColor, SelFgColor Attribute

	// If Editable is true, keystrokes will be added to the view's internal
	// buffer at the cursor position.
	Editable bool

	// Editor allows to define the editor that manages the edition mode,
	// including keybindings or cursor behaviour. DefaultEditor is used by
	// default.
	Editor Editor

	// Overwrite enables or disables the overwrite mode of the view.
	Overwrite bool

	// If Highlight is true, Sel{Bg,Fg}Colors will be used
	// for the line under the cursor position.
	Highlight bool

	// If Frame is true, a border will be drawn around the view.
	Frame bool

	// If Wrap is true, the content that is written to this View is
	// automatically wrapped when it is longer than its width. If true the
	// view's x-origin will be ignored.
	Wrap bool

	// If Autoscroll is true, the View will automatically scroll down when the
	// text overflows. If true the view's y-origin will be ignored.
	Autoscroll bool

	// If Frame is true, Title allows to configure a title for the view.
	Title string

	// If Mask is true, the View will display the mask instead of the real
	// content
	Mask rune
}

type viewLine struct {
	linesX, linesY int // coordinates relative to v.lines
	line           []cell
}

type cell struct {
	chr              rune
	bgColor, fgColor Attribute
}

type lineType []cell

// String returns a string from a given cell slice.
func (l lineType) String() string {
	str := ""
	for _, c := range l {
		str += string(c.chr)
	}
	return str
}

// newView returns a new View object.
func newView(name string, x0, y0, x1, y1 int, mode OutputMode) *View {
	v := &View{
		name:    name,
		x0:      x0,
		y0:      y0,
		x1:      x1,
		y1:      y1,
		visible: true,
		Frame:   true,
		Editor:  DefaultEditor,
		tainted: true,
		ei:      newEscapeInterpreter(mode),
	}
	return v
}

// Size returns the number of visible columns and rows in the View.
func (v *View) Size() (x, y int) {
	return v.x1 - v.x0 - 1, v.y1 - v.y0 - 1
}

// Name returns the name of the view.
func (v *View) Name() string {
	return v.name
}

// setRune sets a rune at the given point relative to the view. It applies the
// specified colors, taking into account if the cell must be highlighted. Also,
// it checks if the position is valid.
func (v *View) setRune(x, y int, ch rune, fgColor, bgColor Attribute) error {
	maxX, maxY := v.Size()
	if x < 0 || x >= maxX || y < 0 || y >= maxY {
		return ErrInvalidPoint
	}

	var (
		ry, rcy int
		err     error
	)
	if v.Highlight {
		_, ry, err = v.realPosition(x, y)
		if err != nil {
			return err
		}
		_, rcy, err = v.realPosition(v.cx, v.cy)
		if err != nil {
			return err
		}
	}

	if v.Mask != 0 {
		fgColor = v.FgColor
		bgColor = v.BgColor
		ch = v.Mask
	} else if v.Highlight && ry == rcy {
		fgColor = v.SelFgColor
		bgColor = v.SelBgColor
	}

	// Don't display NUL characters
	if ch == 0 {
		ch = ' '
	}

	termbox.SetCell(v.x0+x+1, v.y0+y+1, ch,
		termbox.Attribute(fgColor), termbox.Attribute(bgColor))

	return nil
}

// SetCursor sets the cursor position of the view at the given point,
// relative to the view. It checks if the position is valid.
func (v *View) SetCursor(x, y int) error {
	maxX, maxY := v.Size()
	if x < 0 || x >= maxX || y < 0 || y >= maxY {
		return ErrInvalidPoint
	}
	v.cx = x
	v.cy = y
	return nil
}

// Cursor returns the cursor position of the view.
func (v *View) Cursor() (x, y int) {
	return v.cx, v.cy
}

// SetOrigin sets the origin position of the view's internal buffer,
// so the buffer starts to be printed from this point, which means that
// it is linked with the origin point of view. It can be used to
// implement Horizontal and Vertical scrolling with just incrementing
// or decrementing ox and oy.
func (v *View) SetOrigin(x, y int) error {
	if x < 0 || y < 0 {
		return ErrInvalidPoint
	}
	v.ox = x
	v.oy = y
	return nil
}

// Origin returns the origin position of the view.
func (v *View) Origin() (x, y int) {
	return v.ox, v.oy
}

// SetWritePos sets the write position of the view's internal buffer.
// So the next Write call would write directly to the specified position.
func (v *View) SetWritePos(x, y int) error {
	if x < 0 || y < 0 {
		return ErrInvalidPoint
	}
	v.wx = x
	v.wy = y
	return nil
}

// WritePos returns the current write position of the view's internal buffer.
func (v *View) WritePos() (x, y int) {
	return v.wx, v.wy
}

// SetReadPos sets the read position of the view's internal buffer.
// So the next Read call would read from the specified position.
func (v *View) SetReadPos(x, y int) error {
	if x < 0 || y < 0 {
		return ErrInvalidPoint
	}
	v.readBuffer = nil
	v.rx = x
	v.ry = y
	return nil
}

// ReadPos returns the current read position of the view's internal buffer.
func (v *View) ReadPos() (x, y int) {
	return v.rx, v.ry
}

// makeWriteable creates empty cells if required to make position (x, y) writeable.
func (v *View) makeWriteable(x, y int) {
	// TODO: make this more efficient

	// line `y` must be index-able (that's why `<=`)
	for len(v.lines) <= y {
		if cap(v.lines) > len(v.lines) {
			newLen := cap(v.lines)
			if newLen > y {
				newLen = y + 1
			}
			v.lines = v.lines[:newLen]
		} else {
			v.lines = append(v.lines, nil)
		}
	}
	// cell `x` must not be index-able (that's why `<`)
	// append should be used by `lines[y]` user if he wants to write beyond `x`
	for len(v.lines[y]) < x {
		if cap(v.lines[y]) > len(v.lines[y]) {
			newLen := cap(v.lines[y])
			if newLen > x {
				newLen = x
			}
			v.lines[y] = v.lines[y][:newLen]
		} else {
			v.lines[y] = append(v.lines[y], cell{})
		}
	}
}

// writeCells copies []cell to specified location (x, y)
// !!! caller MUST ensure that specified location (x, y) is writeable by calling makeWriteable
func (v *View) writeCells(x, y int, cells []cell) {
	var newLen int
	// use maximum len available
	line := v.lines[y][:cap(v.lines[y])]
	maxCopy := len(line) - x
	if maxCopy < len(cells) {
		copy(line[x:], cells[:maxCopy])
		line = append(line, cells[maxCopy:]...)
		newLen = len(line)
	} else { // maxCopy >= len(cells)
		copy(line[x:], cells)
		newLen = x + len(cells)
		if newLen < len(v.lines[y]) {
			newLen = len(v.lines[y])
		}
	}
	v.lines[y] = line[:newLen]
}

// Write appends a byte slice into the view's internal buffer. Because
// View implements the io.Writer interface, it can be passed as parameter
// of functions like fmt.Fprintf, fmt.Fprintln, io.Copy, etc. Clear must
// be called to clear the view's buffer.
func (v *View) Write(p []byte) (n int, err error) {
	v.tainted = true

	// Fill with empty cells, if writing outside current view buffer
	v.makeWriteable(v.wx, v.wy)

	for _, r := range bytes.Runes(p) {
		switch r {
		case '\n':
			v.wy++
			if v.wy >= len(v.lines) {
				v.lines = append(v.lines, nil)
			}

			fallthrough
			// not valid in every OS, but making runtime OS checks in cycle is bad.
		case '\r':
			v.wx = 0
		default:
			cells := v.parseInput(r)
			if cells == nil {
				continue
			}
			v.writeCells(v.wx, v.wy, cells)
			v.wx += len(cells)
		}
	}

	return len(p), nil
}

// parseInput parses char by char the input written to the View. It returns nil
// while processing ESC sequences. Otherwise, it returns a cell slice that
// contains the processed data.
func (v *View) parseInput(ch rune) []cell {
	cells := []cell{}

	isEscape, err := v.ei.parseOne(ch)
	if err != nil {
		for _, r := range v.ei.runes() {
			c := cell{
				fgColor: v.FgColor,
				bgColor: v.BgColor,
				chr:     r,
			}
			cells = append(cells, c)
		}
		v.ei.reset()
	} else {
		if isEscape {
			return nil
		}
		c := cell{
			fgColor: v.ei.curFgColor,
			bgColor: v.ei.curBgColor,
			chr:     ch,
		}
		cells = append(cells, c)
	}

	return cells
}

// Read reads data into p from the current reading position set by SetReadPos.
// It returns the number of bytes read into p.
// At EOF, err will be io.EOF.
func (v *View) Read(p []byte) (n int, err error) {
	buffer := make([]byte, utf8.UTFMax)
	offset := 0
	if v.readBuffer != nil {
		copy(p, v.readBuffer)
		if len(v.readBuffer) >= len(p) {
			if len(v.readBuffer) > len(p) {
				v.readBuffer = v.readBuffer[len(p):]
			}
			return len(p), nil
		}
		v.readBuffer = nil
	}
	for v.ry < len(v.lines) {
		for v.rx < len(v.lines[v.ry]) {
			count := utf8.EncodeRune(buffer, v.lines[v.ry][v.rx].chr)
			copy(p[offset:], buffer[:count])
			v.rx++
			newOffset := offset + count
			if newOffset >= len(p) {
				if newOffset > len(p) {
					v.readBuffer = buffer[newOffset-len(p):]
				}
				return len(p), nil
			}
			offset += count
		}
		v.rx = 0
		v.ry++
	}
	return offset, io.EOF
}

// Sets read and write pos to (0, 0).
func (v *View) Rewind() {
	v.SetReadPos(0, 0)
	v.SetWritePos(0, 0)
}

// draw re-draws the view's contents.
func (v *View) draw() error {
	if !v.visible {
		return nil
	}

	maxX, maxY := v.Size()

	if v.Wrap {
		if maxX == 0 {
			return errors.New("X size of the view cannot be 0")
		}
		v.ox = 0
	}
	if v.tainted {
		v.viewLines = nil
		for i, line := range v.lines {
			if v.Wrap {
				if len(line) <= maxX {
					vline := viewLine{linesX: 0, linesY: i, line: line}
					v.viewLines = append(v.viewLines, vline)
					continue
				} else {
					vline := viewLine{linesX: 0, linesY: i, line: line[:maxX]}
					v.viewLines = append(v.viewLines, vline)
				}
				// Append remaining lines
				for n := maxX; n < len(line); n += maxX {
					if len(line[n:]) <= maxX {
						vline := viewLine{linesX: n, linesY: i, line: line[n:]}
						v.viewLines = append(v.viewLines, vline)
					} else {
						vline := viewLine{linesX: n, linesY: i, line: line[n : n+maxX]}
						v.viewLines = append(v.viewLines, vline)
					}
				}
			} else {
				vline := viewLine{linesX: 0, linesY: i, line: line}
				v.viewLines = append(v.viewLines, vline)
			}
		}
		v.tainted = false
	}

	if v.Autoscroll && len(v.viewLines) > maxY {
		v.oy = len(v.viewLines) - maxY
	}
	y := 0
	for i, vline := range v.viewLines {
		if i < v.oy {
			continue
		}
		if y >= maxY {
			break
		}
		x := 0
		for j, c := range vline.line {
			if j < v.ox {
				continue
			}
			if x >= maxX {
				break
			}

			fgColor := c.fgColor
			if fgColor == ColorDefault {
				fgColor = v.FgColor
			}
			bgColor := c.bgColor
			if bgColor == ColorDefault {
				bgColor = v.BgColor
			}

			if err := v.setRune(x, y, c.chr, fgColor, bgColor); err != nil {
				return err
			}
			x++
		}
		y++
	}
	return nil
}

// realPosition returns the position in the internal buffer corresponding to the
// point (x, y) of the view.
func (v *View) realPosition(vx, vy int) (x, y int, err error) {
	vx = v.ox + vx
	vy = v.oy + vy

	if vx < 0 || vy < 0 {
		return 0, 0, ErrInvalidPoint
	}

	if len(v.viewLines) == 0 {
		return vx, vy, nil
	}

	if vy < len(v.viewLines) {
		vline := v.viewLines[vy]
		x = vline.linesX + vx
		y = vline.linesY
	} else {
		vline := v.viewLines[len(v.viewLines)-1]
		x = vx
		y = vline.linesY + vy - len(v.viewLines) + 1
	}

	return x, y, nil
}

// Clear empties the view's internal buffer.
// And resets reading and writing offsets.
func (v *View) Clear() {
	v.Rewind()
	v.tainted = true
	v.lines = nil
	v.viewLines = nil
	v.clearRunes()
}

// clearRunes erases all the cells in the view.
func (v *View) clearRunes() {
	maxX, maxY := v.Size()
	for x := 0; x < maxX; x++ {
		for y := 0; y < maxY; y++ {
			termbox.SetCell(v.x0+x+1, v.y0+y+1, ' ',
				termbox.Attribute(v.FgColor), termbox.Attribute(v.BgColor))
		}
	}
}

// Buffer returns a string with the contents of the view's internal
// buffer.
func (v *View) Buffer() string {
	str := ""
	for _, l := range v.lines {
		str += lineType(l).String() + "\n"
	}
	return strings.Replace(str, "\x00", " ", -1)
}

// ViewBuffer returns a string with the contents of the view's buffer that is
// shown to the user.
func (v *View) ViewBuffer() string {
	str := ""
	for _, l := range v.viewLines {
		str += lineType(l.line).String() + "\n"
	}
	return strings.Replace(str, "\x00", " ", -1)
}

// Line returns a string with the line of the view's internal buffer
// at the position corresponding to the point (x, y).
func (v *View) Line(y int) (string, error) {
	_, y, err := v.realPosition(0, y)
	if err != nil {
		return "", err
	}

	if y < 0 || y >= len(v.lines) {
		return "", ErrInvalidPoint
	}

	return lineType(v.lines[y]).String(), nil
}

// Word returns a string with the word of the view's internal buffer
// at the position corresponding to the point (x, y).
func (v *View) Word(x, y int) (string, error) {
	x, y, err := v.realPosition(x, y)
	if err != nil {
		return "", err
	}

	if x < 0 || y < 0 || y >= len(v.lines) || x >= len(v.lines[y]) {
		return "", ErrInvalidPoint
	}

	str := lineType(v.lines[y]).String()

	nl := strings.LastIndexFunc(str[:x], indexFunc)
	if nl == -1 {
		nl = 0
	} else {
		nl = nl + 1
	}
	nr := strings.IndexFunc(str[x:], indexFunc)
	if nr == -1 {
		nr = len(str)
	} else {
		nr = nr + x
	}
	return string(str[nl:nr]), nil
}

// Hide view
func (v *View) Hide() {
	v.visible = false
}

// Show view
func (v *View) Show() {
	v.visible = true
}

// ToggleVisible of the view
func (v *View) ToggleVisible() bool {
	v.visible = !v.visible
	return v.visible
}

// Visible returns the visibility of the view
func (v *View) Visible() bool {
	return v.visible
}

// indexFunc allows to split lines by words taking into account spaces
// and 0.
func indexFunc(r rune) bool {
	return r == ' ' || r == 0
}

// SetLine changes the contents of an existing line.
func (v *View) SetLine(y int, text string) error {
	if y > len(v.lines) {
		err := ErrInvalidPoint
		return err
	}

	v.tainted = true
	line := make([]cell, 0)
	for _, r := range text {
		c := v.parseInput(r)
		line = append(line, c...)
	}
	v.lines[y] = line
	return nil
}

// SetHighlight toggles highlighting of separate lines, for custom lists
// or multiple selection in views.
func (v *View) SetHighlight(y int, on bool) error {
	if y > len(v.lines) {
		err := ErrInvalidPoint
		return err
	}

	line := v.lines[y]
	cells := make([]cell, 0)
	for _, c := range line {
		if on {
			c.bgColor = v.SelBgColor
			c.fgColor = v.SelFgColor
		} else {
			c.bgColor = v.BgColor
			c.fgColor = v.FgColor
		}
		cells = append(cells, c)
	}
	v.tainted = true
	v.lines[y] = cells
	return nil
}
