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

var (
	regexChan   = make(chan string)
	resultsChan = make(chan string)
	fileLog     = log.Logger{}
	userString  string
	userRegex   *regexp.Regexp
	debugMode   bool
	view        string
)

type Colorer func(a ...interface{}) string

func init() {
	flag.BoolVar(&debugMode, "debug", false, "Run in debug mode")
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

	print(line)
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
		case <-regexChan:
			g.Update(func(g *gocui.Gui) error {
				resultsView, err := g.View("results")
				regexView, err := g.View("regex")
				errorsView, err := g.View("errors")

				if err != nil {
					return err
				}

				errorsView.Clear()

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
					resultsView.Clear()
					fmt.Fprint(resultsView, userString)
					errorsView.Clear()
					fmt.Fprint(errorsView, err.Error())
					return nil
				}

				userRegex = re

				matches := ReturnsMatch(re, userString)
				resultsView.Clear()
				PrintResults(resultsView, userString, matches)
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
				PrintResults(resultsView, userString, matches)
				return nil
			})
		}
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if regexView, err := g.SetView("regex", 1, 1, maxX-2, 3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		regexView.Editable = true
		regexView.Editor = gocui.EditorFunc(regexEditor)
		regexView.Wrap = true
	}

	if resultsView, err := g.SetView("results", 1, 4, maxX-2, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		fmt.Fprint(resultsView, userString)

		resultsView.Editable = true
		resultsView.Editor = gocui.EditorFunc(resultsEditor)
		resultsView.Wrap = true
		g.SetCurrentView("results")
	}

	if _, err := g.SetView("errors", 1, maxY-4, maxX-2, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
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

	go func() { regexChan <- "" }()
}

func quit(g *gocui.Gui, v *gocui.View) error {
	close(regexChan)
	close(resultsChan)
	return gocui.ErrQuit
}

func PrintResults(w io.Writer, userString string, matches [][]int) {
	// Setup the color function
	blue := color.New(color.BgBlue).SprintFunc()
	red := color.New(color.BgRed).SprintFunc()

  lastIndex := 0
  for index, matchTuple := range matches {
		var colorer Colorer

    if index%2 == 0 {
      colorer = red
    } else {
      colorer = blue
    }

    startMatch := matchTuple[0]
    endMatch := matchTuple[1]

    var prefixSlice string
    if startMatch > 0 {
      prefixSlice = userString[lastIndex:startMatch]
    }

    highlightSlice := userString[startMatch:endMatch]
    lastIndex = endMatch

    fmt.Fprintf(w, "%s", string(prefixSlice))
    fmt.Fprintf(w, "%s", colorer(string(highlightSlice)))
  }
  fmt.Fprintf(w, "%s", string(userString[lastIndex:]))
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
