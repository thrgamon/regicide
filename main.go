package main

import (
	"fmt"
	"io"
	"log"
	_ "os"
	"regexp"

	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
)

func main() {
	// userRegex := os.Args[1]
	// userString := os.Args[2]
	// re := regexp.MustCompile(userRegex)
	// results := ReturnsMatch(re, userString)
	// for _, result := range results {
	//   PrintResults(os.Stdout, userString, result)
	// }

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
	g.Mouse = true
	g.Cursor = true

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, _ := g.Size()
	if v, err := g.SetView("regex", 1, 1, maxX-2, 3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = true
		var DefaultEditor gocui.Editor = gocui.EditorFunc(simpleEditor)
		v.Editor = DefaultEditor
		v.Wrap = true
		g.SetCurrentView("regex")
	}

	return nil
}

func simpleEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyArrowDown:
		v.MoveCursor(0, 1, false)
	case key == gocui.KeyArrowUp:
		v.MoveCursor(0, -1, false)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func PrintResults(w io.Writer, userString string, matchIndex []int) {
	// Get the beginning and the end of the match
	ms := matchIndex[0]
	me := matchIndex[1]

	// Split the string around the matches
	prefix := userString[:ms]
	result := userString[ms:me]
	suffix := userString[me:]

	// Setup the color function
	red := color.New(color.FgRed).SprintFunc()

	// Print, highlighting the match
	fmt.Fprintf(w, "%s%s%s\n", prefix, red(result), suffix)
}

func ReturnsMatch(re *regexp.Regexp, comparitor string) (results [][]int) {
	ba := []byte(comparitor)
	return re.FindAllIndex(ba, -1)
}
