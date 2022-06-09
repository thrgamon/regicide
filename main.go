package main

import (
	"flag"
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
var debugMode bool
var multiLine bool

func init() {
	flag.BoolVar(&debugMode, "debug", false, "Run in debug mode")
	flag.BoolVar(&multiLine, "multiline", false, "Run against multiple lines of text")
}

func main() {
	flag.Parse()

	if debugMode {
    f := sendLogsToFile()
    defer f.Close()
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}

	g.SetManagerFunc(layout)
	g.Cursor = true

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	go updater(g)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

	rv, _ := g.View("regex")
	line := rv.ViewBuffer()
	g.Close()
	println(line)
}

func updater(g *gocui.Gui) {
	for {
		select {
		case reg := <-regex:
			g.Update(func(g *gocui.Gui) error {
				v, err := g.View("results")
				rv, err := g.View("regex")

				if err != nil {
					return err
				}

				if debugMode {
					fileLog.Print(reg)
				}

				v.Clear()
				reRaw := strings.Replace(rv.ViewBuffer(), "\n", "", 1)
				if reRaw == "" {
					fmt.Fprint(v, userString)
					return nil
				}

				re, err := regexp.Compile(reRaw)
				if err != nil {
					fmt.Fprint(v, err.Error())
				} else {
					matches := ReturnsMatch(re, userString)
					if true {
						PrintResultsMultiline(v, userString, matches)
					} else {
						for _, result := range matches {
							PrintResults(v, userString, result)
						}
					}
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
		v.Editor = gocui.EditorFunc(simpleEditor)
		v.Wrap = true
		g.SetCurrentView("regex")
	}

	if vr, err := g.SetView("results", 1, 4, maxX-2, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprint(vr, userString)
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

	go func() { regex <- "" }()
}

func quit(g *gocui.Gui, v *gocui.View) error {
	close(regex)
	return gocui.ErrQuit
}

type Colorer func(a ...interface{}) string

func PrintResultsMultiline(w io.Writer, userString string, matches [][]int) {
	// Setup the color function
	blue := color.New(color.BgBlue).SprintFunc()
	red := color.New(color.BgRed).SprintFunc()

	for index, char := range userString {
		var match bool
		var colorer Colorer

		for mi, matchIndex := range matches {
			if mi%2 == 0 {
				colorer = red
			} else {
				colorer = blue
			}

			if index >= matchIndex[0] && index < matchIndex[1] {
				match = true
			}
			if match == true {
				break
			}
		}

		if match == true {
			fmt.Fprintf(w, "%s", colorer(string(char)))
		} else {
			fmt.Fprintf(w, "%s", string(char))
		}
	}
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

func sendLogsToFile() *os.File {
  f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
  if err != nil {
    log.Fatalf("error opening file: %v", err)
  }
  fileLog.SetOutput(f)

  return f
}

