// Demo code for the TreeView primitive.
package main

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app     *tview.Application
	details *tview.TextArea
	grid    *tview.Grid
	header  *tview.TextView
	pager   *tview.Pages
	tree    *tview.TreeView
)

// Show a navigable tree view of the current directory.
func main() {
	InitLog()
	header = textView("header")
	dbfile := "test.db"
	if len(os.Args) == 2 {
		dbfile = os.Args[1]
	}
	if err := InitDatabase(dbfile); err != nil {
		panic(err)
	}
	details = textArea("details")
	details.SetBorder(true).SetTitle("Details").SetTitleAlign(tview.AlignCenter)
	tree = newTree(details)

	grid = mainGrid()

	pager = tview.NewPages().AddPage("main", grid, true, true)
	pager.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		log.Println("pager event handler", event.Key())
		if event.Key() == tcell.KeyEsc {
			log.Println("pager is handling esc")
			front, _ := pager.GetFrontPage()
			if front != "main" {
				pager.RemovePage(front)
				return nil
			}
		}
		return event
	})

	app = tview.NewApplication().SetRoot(pager, true).EnableMouse(true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		log.Println(event.Key(), event.Rune(), event.Modifiers())
		switch event.Key() {
		case tcell.KeyF1:
			help := helpView()
			pager.AddPage("help", help, true, true)
			app.SetFocus(help)
			return nil
		case tcell.KeyF2:
			front, item := pager.GetFrontPage()
			log.Println("pages", pager.GetPageNames(false), pager.GetPageNames(true), front, item)
		case tcell.KeyCtrlQ:
			app.Stop()
		case tcell.KeyCtrlC:
			return nil
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
	logFile, err := os.CreateTemp("", "bblotEdit*.log")
	if err != nil {
		log.Fatal(err)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Default().SetOutput(logFile)
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

func mainGrid() *tview.Grid {
	grid = tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0, 0).
		SetBorders(true).
		AddItem(header, 0, 0, 1, 2, 0, 0, false).
		AddItem(textView("press ? for help, esc or ctrl-Q to quit"), 2, 0, 1, 2, 0, 0, false).
		AddItem(tree, 1, 0, 1, 1, 0, 0, true).
		AddItem(details, 1, 1, 1, 1, 0, 0, false)
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		log.Println("grid key handler", event.Key())
		return event
	})
	return grid
}
