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
var resultsChan = make(chan string)
var fileLog = log.Logger{}
var userString string
var userRegex *regexp.Regexp
var debugMode bool
var multiLine bool
var view string

func init() {
	flag.BoolVar(&debugMode, "debug", false, "Run in debug mode")
	flag.BoolVar(&multiLine, "multiline", false, "Run against multiple lines of text")
}

func main() {
	flag.Parse()
  userString = flag.Arg(0)

	if debugMode {
		f := sendLogsToFile()
		defer f.Close()
	}

	g := setupGui()

	go updater(g)

	startMainLoop(g)

	closeDownGui(g)
}

func startMainLoop(g *gocui.Gui) {
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func closeDownGui(g *gocui.Gui) {
	line := fetchCurrentRegex(g)

	g.Close()

	println(line)
}

func setupGui() *gocui.Gui {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	g.SetManagerFunc(layout)
	g.Cursor = true

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, changeView); err != nil {
		log.Panicln(err)
	}

	return g
}

func changeView(g *gocui.Gui, v *gocui.View) error {
  if g.CurrentView().Name() == "results" {
	  g.SetCurrentView("regex")
  } else {
	  g.SetCurrentView("results")
  }
  return nil
}

func fetchCurrentRegex(g *gocui.Gui) string {
	rv, _ := g.View("regex")
	line := rv.ViewBuffer()
	return line
}

func updater(g *gocui.Gui) {
	for {
		select {
		case <-regex:
			g.Update(func(g *gocui.Gui) error {
				resultsView, err := g.View("results")
				regexView, err := g.View("regex")

				if err != nil {
					return err
				}

				reRaw := strings.Replace(regexView.ViewBuffer(), "\n", "", 1)

				// If regex is an empty string then just print
				// the plain user input with no matching
				if reRaw == "" {
				  resultsView.Clear()
					fmt.Fprint(resultsView, userString)
					return nil
				}

				re, err := regexp.Compile(reRaw)
				if err != nil {
					fmt.Fprint(resultsView, err.Error())
					return nil
				}

        userRegex = re

				matches := ReturnsMatch(re, userString)
				resultsView.Clear()
				if multiLine == true {
					PrintResultsMultiline(resultsView, userString, matches)
				} else {
					for _, result := range matches {
						PrintResults(resultsView, userString, result)
					}
				}

				return nil
			})
		case <-resultsChan:
			g.Update(func(g *gocui.Gui) error {
				resultsView, err := g.View("results")

				if err != nil {
					return err
				}

        userString = resultsView.ViewBuffer()

        if userRegex == nil {
				  resultsView.Clear()
					fmt.Fprint(resultsView, userString)
          return nil
        }

				matches := ReturnsMatch(userRegex, userString)
				resultsView.Clear()
				if multiLine == true {
					PrintResultsMultiline(resultsView, userString, matches)
				} else {
					for _, result := range matches {
						PrintResults(resultsView, userString, result)
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
		v.Editor = gocui.EditorFunc(regexEditor)
		v.Wrap = true
	}

	if vr, err := g.SetView("results", 1, 4, maxX-2, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprint(vr, userString)

		vr.Editable = true
		vr.Editor = gocui.EditorFunc(resultsEditor)
		vr.Wrap = true
		g.SetCurrentView("results")
	}

	return nil
}


func resultsEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyEnter:
		v.EditNewLine()
	case key == gocui.KeyArrowDown:
		v.MoveCursor(0, 1, false)
	case key == gocui.KeyArrowUp:
		v.MoveCursor(0, -1, false)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	}

	go func() { resultsChan <- "" }()
}

func regexEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
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
