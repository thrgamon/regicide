package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
)

var regex = make(chan string)
var fileLog = log.Logger{}
var userString = os.Args[1]

func main() {
  f, err := os.OpenFile("testlogfile", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  if err != nil {
      log.Fatalf("error opening file: %v", err)
  }
  defer f.Close()

  fileLog.SetOutput(f)
	// userRegex := os.Args[1]
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


  go updater(g)


	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func updater(g *gocui.Gui) {
  for {
    select {
  case reg := <-regex:
    g.Update(func(g *gocui.Gui) error {
      v,err := g.View("results")
      rv,err := g.View("regex")

      if err != nil {
        return err
      }

      fileLog.Print(reg)

      v.Clear()
      re := regexp.MustCompile(strings.Replace(rv.ViewBuffer(), "\n", "", 1))
      matches := ReturnsMatch(re, userString)
      for _, result := range matches {
        PrintResults(v,userString, result)
      }
      return nil
    })
  }

  }
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
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

	if _, err := g.SetView("results", 1, 4, maxX-2, maxY - 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
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

  var buf []byte
  v.Read(buf)

  go func() {regex <- string(buf)}()
}

func quit(g *gocui.Gui, v *gocui.View) error {
  close(regex)
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
