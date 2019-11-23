// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import (
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/go-errors/errors"
)

type escapeInterpreter struct {
	state                  escapeState
	curch                  rune
	csiParam               []string
	curFgColor, curBgColor tcell.Color
	curAttr                tcell.AttrMask
}

type (
	escapeState int
	fontEffect  int
)

const (
	stateNone escapeState = iota
	stateEscape
	stateCSI
	stateParams

	bold               fontEffect = 1
	underline          fontEffect = 4
	reverse            fontEffect = 7
	setForegroundColor fontEffect = 38
	setBackgroundColor fontEffect = 48
)

var (
	errNotCSI        = errors.New("Not a CSI escape sequence")
	errCSIParseError = errors.New("CSI escape sequence parsing error")
	errCSITooLong    = errors.New("CSI escape sequence is too long")
)

// runes in case of error will output the non-parsed runes as a string.
func (ei *escapeInterpreter) runes() []rune {
	switch ei.state {
	case stateNone:
		return []rune{0x1b}
	case stateEscape:
		return []rune{0x1b, ei.curch}
	case stateCSI:
		return []rune{0x1b, '[', ei.curch}
	case stateParams:
		ret := []rune{0x1b, '['}
		for _, s := range ei.csiParam {
			ret = append(ret, []rune(s)...)
			ret = append(ret, ';')
		}
		return append(ret, ei.curch)
	}
	return nil
}

// newEscapeInterpreter returns an escapeInterpreter that will be able to parse
// terminal escape sequences.
func newEscapeInterpreter() *escapeInterpreter {
	ei := &escapeInterpreter{
		state:      stateNone,
		curFgColor: tcell.ColorDefault,
		curBgColor: tcell.ColorDefault,
		curAttr:    tcell.AttrNone,
	}
	return ei
}

// reset sets the escapeInterpreter in initial state.
func (ei *escapeInterpreter) reset() {
	ei.state = stateNone
	ei.curFgColor = tcell.ColorDefault
	ei.curBgColor = tcell.ColorDefault
	ei.curAttr = tcell.AttrNone
	ei.csiParam = nil
}

// parseOne parses a rune. If isEscape is true, it means that the rune is part
// of an escape sequence, and as such should not be printed verbatim. Otherwise,
// it's not an escape sequence.
func (ei *escapeInterpreter) parseOne(ch rune) (isEscape bool, err error) {
	// Sanity checks
	if len(ei.csiParam) > 20 {
		return false, errCSITooLong
	}
	if len(ei.csiParam) > 0 && len(ei.csiParam[len(ei.csiParam)-1]) > 255 {
		return false, errCSITooLong
	}

	ei.curch = ch

	switch ei.state {
	case stateNone:
		if ch == 0x1b {
			ei.state = stateEscape
			return true, nil
		}
		return false, nil
	case stateEscape:
		if ch == '[' {
			ei.state = stateCSI
			return true, nil
		}
		return false, errNotCSI
	case stateCSI:
		switch {
		case ch >= '0' && ch <= '9':
			ei.csiParam = append(ei.csiParam, "")
		case ch == 'm':
			ei.csiParam = append(ei.csiParam, "0")
		default:
			return false, errCSIParseError
		}
		ei.state = stateParams
		fallthrough
	case stateParams:
		switch {
		case ch >= '0' && ch <= '9':
			ei.csiParam[len(ei.csiParam)-1] += string(ch)
			return true, nil
		case ch == ';':
			ei.csiParam = append(ei.csiParam, "")
			return true, nil
		case ch == 'm':
			err := ei.output256()
			if err != nil {
				return false, errCSIParseError
			}

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil
		default:
			return false, errCSIParseError
		}
	}
	return false, nil
}

// outputNormal provides 8 different colors:
//   black, red, green, yellow, blue, magenta, cyan, white
func (ei *escapeInterpreter) outputNormal() error {
	for _, param := range ei.csiParam {
		p, err := strconv.Atoi(param)
		if err != nil {
			return errCSIParseError
		}

		switch {
		case p >= 30 && p <= 37:
			ei.curFgColor = tcell.Color(p - 30)
		case p == 39:
			ei.curFgColor = tcell.ColorDefault
		case p >= 40 && p <= 47:
			ei.curBgColor = tcell.Color(p - 40)
		case p == 49:
			ei.curBgColor = tcell.ColorDefault
		case p == 1:
			ei.curAttr |= tcell.AttrBold
		case p == 4:
			ei.curAttr |= tcell.AttrUnderline
		case p == 7:
			ei.curAttr |= tcell.AttrReverse
		case p == 0:
			ei.curFgColor = tcell.ColorDefault
			ei.curBgColor = tcell.ColorDefault
			ei.curAttr = tcell.AttrNone
		}
	}

	return nil
}

// output256 allows you to leverage the 256-colors terminal mode:
//   0x01 - 0x08: the 8 colors as in OutputNormal
//   0x09 - 0x10: Color* | AttrBold
//   0x11 - 0xe8: 216 different colors
//   0xe9 - 0x1ff: 24 different shades of grey
func (ei *escapeInterpreter) output256() error {
	if len(ei.csiParam) < 3 {
		return ei.outputNormal()
	}

	mode, err := strconv.Atoi(ei.csiParam[1])
	if err != nil {
		return errCSIParseError
	}
	if mode != 5 {
		return ei.outputNormal()
	}

	for _, param := range splitFgBg(ei.csiParam) {
		fgbg, err := strconv.Atoi(param[0])
		if err != nil {
			return errCSIParseError
		}
		color, err := strconv.Atoi(param[2])
		if err != nil {
			return errCSIParseError
		}

		switch fontEffect(fgbg) {
		case setForegroundColor:
			ei.curFgColor = tcell.Color(color)

			for _, s := range param[3:] {
				p, err := strconv.Atoi(s)
				if err != nil {
					return errCSIParseError
				}

				switch fontEffect(p) {
				case bold:
					ei.curAttr |= tcell.AttrBold
				case underline:
					ei.curAttr |= tcell.AttrUnderline
				case reverse:
					ei.curAttr |= tcell.AttrReverse
				}
			}
		case setBackgroundColor:
			ei.curBgColor = tcell.Color(color)
		default:
			return errCSIParseError
		}
	}
	return nil
}

func splitFgBg(params []string) [][]string {
	var out [][]string
	var current []string
	for _, p := range params {
		if len(current) == 3 && (p == "48" || p == "38") {
			out = append(out, current)
			current = []string{}
		}
		current = append(current, p)
	}

	if len(current) > 0 {
		out = append(out, current)
	}

	return out
}
