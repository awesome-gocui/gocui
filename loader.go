package gocui

import (
	"time"

	"github.com/gdamore/tcell"
)

func (g *Gui) loaderTick() {
	go func() {
		for range time.Tick(time.Millisecond * 50) {
			for _, view := range g.Views() {
				if view.HasLoader {
					g.userEvents <- userEvent{func(g *Gui) error { return nil }}
					break
				}
			}
		}
	}()
}

func (v *View) loaderLines() [][]cell {
	duplicate := make([][]cell, len(v.lines))
	for i := range v.lines {
		if i < len(v.lines)-1 {
			duplicate[i] = newCellArray(len(v.lines[i]))
			copy(duplicate[i], v.lines[i])
		} else {
			duplicate[i] = newCellArray(len(v.lines[i]) + 2)
			copy(duplicate[i], v.lines[i])
			duplicate[i][len(duplicate[i])-2] = cell{
				chr:     ' ',
				bgColor: tcell.ColorDefault,
				fgColor: tcell.ColorDefault,
			}
			duplicate[i][len(duplicate[i])-1] = Loader()
		}
	}

	return duplicate
}

// Loader can show a loading animation
func Loader() cell {
	characters := "|/-\\"
	now := time.Now()
	nanos := now.UnixNano()
	index := nanos / 50000000 % int64(len(characters))
	str := characters[index : index+1]
	chr := []rune(str)[0]
	return cell{
		chr:     chr,
		bgColor: tcell.ColorDefault,
		fgColor: tcell.ColorDefault,
	}
}
