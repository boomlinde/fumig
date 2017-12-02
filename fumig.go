package main

import (
	"errors"
	"fmt"
	tb "github.com/nsf/termbox-go"
	"github.com/skratchdot/open-golang/open"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type hEntry struct {
	selector *gopherline
	offset   int
}

var textCache = make(map[string][]string)
var menuCache = map[string][]*gopherline{"gopher://__start__:0/1": helpPage}
var history = []*hEntry{}
var status string
var hasError bool

func enter(selector *gopherline) {
	history = append(history, &hEntry{selector, 0})
}

func back() {
	if len(history) < 2 {
		return
	}
	history = history[:len(history)-1]
}

func isMenu(ftype rune) bool {
	return ftype == '1' || ftype == '7'
}

func get(selector *gopherline) (menu []*gopherline, text []string, err error) {
	var ok bool
	uri := selector.ToUri()
	switch selector.Ftype {
	case '7':
		fallthrough
	case '1':
		menu, ok = menuCache[uri]
		if !ok {
			menu, err = getMenu(selector)
			if err != nil {
				return
			}
			menuCache[uri] = menu
		}
	case '0':
		text, ok = textCache[uri]
		if !ok {
			text, err = getPlain(selector)
			if err != nil {
				return
			}
			textCache[uri] = text
		}
	default:
		err = errors.New("No handler yet")
	}

	return
}

func drawStr(x, y int, s string, fg, bg tb.Attribute) {
	i := 0
	for _, r := range s {
		tb.SetCell(x+i, y, r, fg, bg)
		i++
	}
}

func size() (int, int) {
	width, height := tb.Size()
	return width, height - 1
}

func prompt(question string) bool {
	question += " (y/n)"
	redraw := func() {
		width, height := tb.Size()
		for i := 0; i < width; i++ {
			tb.SetCell(i, height-1, ' ', tb.ColorDefault, tb.ColorDefault)
		}
		drawStr(0, height-1, question, tb.ColorDefault, tb.ColorDefault)
		tb.Flush()
	}

	for {
		redraw()
		e := tb.PollEvent()
		if e.Type != tb.EventKey {
			continue
		}
		if e.Ch == 'Y' || e.Ch == 'y' {
			return true
		} else if e.Ch == 'N' || e.Ch == 'n' {
			return false
		}
	}

}
func editbox(prompt string, init string) (string, bool) {
	defer tb.HideCursor()
	buf := []rune(init)
	index := len(buf)
	redraw := func() {
		width, height := tb.Size()
		for i := 0; i < width; i++ {
			tb.SetCell(i, height-1, ' ', tb.ColorDefault, tb.ColorDefault)
		}
		drawStr(0, height-1, prompt, tb.ColorDefault, tb.ColorDefault)
		drawStr(len(prompt), height-1, string(buf),
			tb.ColorDefault, tb.ColorDefault)
		tb.SetCursor(len(prompt)+index, height-1)
		tb.Flush()
	}

	for {
		redraw()
		e := tb.PollEvent()
		if e.Type != tb.EventKey {
			continue
		}
		if e.Ch != 0 {
			buf = append(buf[:index], append([]rune{e.Ch}, buf[index:]...)...)
			index++
		} else {
			switch e.Key {
			case tb.KeyEnter:
				return string(buf), true
			case tb.KeyEsc:
				return string(buf), false
			case tb.KeySpace:
				buf = append(buf[:index], append([]rune{' '}, buf[index:]...)...)
				index++
			case tb.KeyArrowLeft:
				if index > 0 {
					index--
				}
			case tb.KeyArrowRight:
				if index < len(buf) {
					index++
				}
			case tb.KeyBackspace:
				fallthrough
			case tb.KeyBackspace2:
				if index > 0 {
					buf = append(buf[:index-1], buf[index:]...)
					index--
				}
			case tb.KeyDelete:
				if index < len(buf) {
					buf = append(buf[:index], buf[index+1:]...)
				}
			case tb.KeyHome:
				index = 0
			case tb.KeyEnd:
				index = len(buf)
			}
		}
	}
}

func draw(menu []*gopherline, text []string) {
	var current *hEntry
	width, height := size()
	tb.Clear(tb.ColorDefault, tb.ColorDefault)
	if len(history) != 0 {
		current = history[len(history)-1]
	} else {
		current = &hEntry{&gopherline{}, 0}
	}
	var title string
	title = current.selector.Text
	drawStr(0, height, title, tb.ColorYellow|tb.AttrBold, tb.ColorDefault)
	if hasError {
		drawStr(width-len(status), height, status, tb.ColorRed|tb.AttrBold, tb.ColorDefault)
		hasError = false
	} else {
		drawStr(width-len(status), height, status, tb.ColorCyan|tb.AttrBold, tb.ColorDefault)
	}
	var offset int
	switch current.selector.Ftype {
	case '7':
		fallthrough
	case '1':
		offset = current.offset - height/2
		if offset+height > len(menu) {
			offset = len(menu) - height
		}
		if offset < 0 {
			offset = 0
		}
		for i := offset; i < len(menu) && i-offset < height; i++ {
			fg := tb.ColorGreen
			fgt := tb.ColorDefault
			if menu[i].Ftype == 'i' {
				fg = tb.ColorWhite
			}
			if i == current.offset {
				fg |= tb.AttrBold
				fgt |= tb.AttrBold
				drawStr(0, i-offset, ">", fg, tb.ColorDefault)
			}
			nice := menu[i].NiceType()
			if strings.HasPrefix(menu[i].Path, "URL:") {
				nice = "(WWW)"
			}
			drawStr(2, i-offset, nice, fgt, tb.ColorDefault)
			drawStr(8, i-offset, menu[i].Text, fg, tb.ColorDefault)
		}
	case '0':
		offset = current.offset
		if offset+height > len(text) {
			offset = len(text) - height
		}
		if offset < 0 {
			offset = 0
		}
		current.offset = offset
		for i := offset; i < len(text) && i-offset < height; i++ {
			drawStr(0, i-offset, text[i], tb.ColorDefault, tb.ColorDefault)
		}
	}
	tb.Flush()
}

func updateOffset(menu []*gopherline, text []string, add int) {
	current := history[len(history)-1]
	current.offset += add
	switch current.selector.Ftype {
	case '7':
		fallthrough
	case '1':
		if current.offset >= len(menu) {
			current.offset = len(menu) - 1
		}
	case '0':
		if current.offset >= len(text) {
			current.offset = len(text) - 1
		}
	}
	if current.offset < 0 {
		current.offset = 0
	}
}

func main() {
	var err error
	var menu []*gopherline
	var text []string
	var start string

	tmpdir, err := ioutil.TempDir("", "fumig")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create temp dir\n")
		os.Exit(1)
	}
	defer os.RemoveAll(tmpdir)

	err = tb.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
	defer tb.Close()

	selector := &gopherline{}

	if len(os.Args) >= 2 {
		start = os.Args[1]
	} else {
		start = "gopher://__start__:0/1"
	}
	err = selector.FromUri(start)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
	selector.Text = "Start"

	status = "Loading..."
	draw(menu, text)

	enter(selector)

	_, height := size()
	for {
		current := history[len(history)-1]
		oldmenu, oldtext := menu, text
		menu, text, err = get(current.selector)
		if err != nil {
			menu, text = oldmenu, oldtext
			back()
			current = history[len(history)-1]
			hasError = true
			status = "Error: " + err.Error()
		} else if !hasError {
			status = current.selector.ToUri()
			if isMenu(current.selector.Ftype) && menu[current.offset].Ftype != 'i' {
				status = menu[current.offset].ToUri()
				if menu[current.offset].Ftype == 'h' && strings.HasPrefix(menu[current.offset].Path, "URL:") {
					status = menu[current.offset].Path[len("URL:"):]
				}
			}
			if strings.HasPrefix(status, "gopher://") {
				status = status[len("gopher://"):]
			}
		}
		updateOffset(menu, text, 0)
		draw(menu, text)
		event := tb.PollEvent()
		if event.Type == tb.EventResize {
			_, height = size()
			draw(menu, text)
		}

		ke := getKey(event)
		switch ke {
		case keyQuit:
			if prompt("Really quit?") {
				return
			}
		case keyBack:
			back()
		case keyDown:
			updateOffset(menu, text, 1)
		case keyUp:
			updateOffset(menu, text, -1)
		case keyNext:
			if isMenu(current.selector.Ftype) {
				orig := current.offset
				found := false
				for {
					updateOffset(menu, text, 1)
					if menu[current.offset].Ftype != 'i' {
						found = true
						break
					}
					if current.offset == len(menu)-1 {
						break
					}
				}
				if !found {
					current.offset = orig
				}
			}
		case keyPrev:
			if isMenu(current.selector.Ftype) {
				orig := current.offset
				found := false
				for {
					updateOffset(menu, text, -1)
					if menu[current.offset].Ftype != 'i' {
						found = true
						break
					}
					if current.offset == 0 {
						break
					}
				}
				if !found {
					current.offset = orig
				}
			}
		case keyPgup:
			updateOffset(menu, text, -height/2)
		case keyPgdn:
			updateOffset(menu, text, height/2)
		case keyUrl:
			fallthrough
		case keyEnter:
			e := &gopherline{}
			if ke == keyEnter && !isMenu(current.selector.Ftype) {
				break
			}
			if ke != keyEnter {
				if url, ok := editbox("Address: ", ""); ok {
					if err := e.FromUri(url); err != nil {
						hasError = true
						status = err.Error()
						break
					}
				} else {
					break
				}
			} else {
				e = menu[current.offset]
			}

			switch e.Ftype {
			case 'i':
				// do nothing
			case '7':
				if query, ok := editbox("Query: ", ""); ok {
					status = "Loading..."
					draw(menu, text)
					enter(e.copy())
					current = history[len(history)-1]
					current.selector.Path += "\t" + query
				}
			case '1':
				fallthrough
			case '0':
				status = "Loading..."
				draw(menu, text)
				enter(e)
			default:
				// Don't try to download web links
				if e.Ftype == 'h' && strings.HasPrefix(e.Path, "URL:") {
					break
				}
				fname := path.Base(strings.Replace(e.Path, "\\", "/", -1))
				if fname == "/" || fname == "" {
					fname = "unknown.bin"
				}
				if fpath, ok := editbox("Save as: ", fname); ok {
					status = "Downloading " + fpath + "..."
					draw(menu, text)
					err := downloadFile(fpath, e)
					if err != nil {
						status = err.Error()
						hasError = true
					}
				}
			}
		case keyReload:
			uri := current.selector.ToUri()
			switch current.selector.Ftype {
			case '7':
				fallthrough
			case '1':
				if uri == "gopher://__start__:0/1" {
					break
				}
				status = "Reloading..."
				draw(menu, text)
				delete(menuCache, uri)
			case '0':
				status = "Reloading..."
				draw(menu, text)
				delete(textCache, uri)
			}
		case keyTop:
			current.offset = 0
		case keyBottom:
			switch current.selector.Ftype {
			case '7':
				fallthrough
			case '1':
				current.offset = len(menu) - 1
			case '0':
				current.offset = len(text) - 1
			}
		case keyRedraw:
			tb.Sync()
		case keyOpen:
			var fname string
			if !isMenu(current.selector.Ftype) {
				break
			}
			sel := menu[current.offset]
			if strings.ContainsRune("i17", sel.Ftype) {
				break
			}
			if sel.Ftype == 'h' && strings.HasPrefix(sel.Path, "URL:") {
				fname = sel.Path[len("URL:"):]
			} else {
				fname = path.Base(strings.Replace(sel.Path, "\\", "/", -1))
			}
			if sel.Ftype == 'h' && strings.HasPrefix(sel.Path, "URL:") {
				if err := open.Start(fname); err != nil {
					status = err.Error()
					hasError = true
					break
				}
				break
			}
			status = "Loading..."
			draw(menu, text)

			fpath := path.Join(tmpdir, fname)
			if err := downloadFile(fpath, sel); err != nil {
				status = err.Error()
				hasError = true
				break
			}

			if err := open.Start(fpath); err != nil {
				status = err.Error()
				hasError = true
				break
			}
		}
	}
}
