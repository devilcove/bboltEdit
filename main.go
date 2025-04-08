// Demo code for the TreeView primitive.
package main

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lmittmann/tint"
	"github.com/rivo/tview"
)

var (
	app     *tview.Application
	errDisp *tview.Modal
	header  *tview.TextView
	pager   *tview.Pages
	tree    *tview.TreeView
)

// Show a navigable tree view of the current directory.
func main() {
	InitLog()
	header = textView("header")
	if err := InitDatabase("test.db"); err != nil {
		panic(err)
	}

	help := help()
	errDisp = errorView()
	details := textArea("details")
	details.SetBorder(true).SetTitle("Details").SetTitleAlign(tview.AlignCenter)
	tree = newTree(details)
	file := newFiles()

	grid := tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0, 0).
		SetBorders(true).
		AddItem(header, 0, 0, 1, 2, 0, 0, false).
		AddItem(textView("press alt-? for help, ctrl-Q to quit"), 2, 0, 1, 2, 0, 0, false).
		AddItem(tree, 1, 0, 1, 1, 0, 0, true).
		AddItem(details, 1, 1, 1, 1, 0, 0, false)
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		log.Println("grid key handler", event.Key())
		return event
	})

	pager = tview.NewPages().
		AddPage("main", grid, true, true).
		AddPage("help", help, true, false).
		AddPage("file", file, true, false).
		AddPage("error", errDisp, true, false)
	pager.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		log.Println("pager event handler", event.Key())
		return event
	})

	help.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pager.HidePage("help")
	})
	errDisp.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pager.HidePage("error")
	})

	app = tview.NewApplication().SetRoot(pager, true).EnableMouse(true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		log.Println(event.Key(), event.Rune(), event.Modifiers())
		switch event.Key() {
		case tcell.KeyEscape:
			log.Println("esc key")
			pager.SwitchToPage("main")
			pages := pager.GetPageNames(false)
			log.Println(pages)
		case tcell.KeyCtrlQ:
			app.Stop()
		case tcell.KeyCtrlC:
			return nil
		}
		if event.Modifiers() == tcell.ModAlt {
			log.Println("alt modifier")
			switch event.Rune() {
			case '?':
				pager.ShowPage("help")
				app.SetFocus(help)
			case 'o':
				pager.SwitchToPage("file")
				app.SetFocus(file)
			}
		}
		log.Println("app key handling: passing ", event.Key())
		return event
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}

// InitLog creates a file to use for debugging messages
func InitLog() {
	debugFile, err := os.OpenFile("debug.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	logLevel := &slog.LevelVar{}
	logger := slog.New(tint.NewHandler(debugFile, &tint.Options{
		AddSource:  true,
		Level:      logLevel,
		TimeFormat: time.Kitchen,
		NoColor:    true,
	}))
	slog.SetDefault(logger)
	// Writer = debugFile
}

func textArea(text string) *tview.TextArea {
	return tview.NewTextArea().
		SetText(text, true)
}

func textView(text string) *tview.TextView {
	return tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(text)
}
