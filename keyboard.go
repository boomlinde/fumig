package main

import tb "github.com/nsf/termbox-go"

const (
	keyUp = iota
	keyDown
	keyBack
	keyEnter
	keyQuit
	keyReload
	keyNext
	keyPrev
	keyPgup
	keyPgdn
	keyTop
	keyBottom
	keyRedraw
	keyOpen
	keyUrl
	keyDownload
)

var chmap = map[rune]int{
	'k': keyUp,
	'j': keyDown,
	'h': keyBack,
	'l': keyEnter,
	'u': keyPgup,
	'd': keyPgdn,
	'r': keyReload,
	'J': keyNext,
	'K': keyPrev,
	'n': keyNext,
	'p': keyPrev,
	'q': keyQuit,
	'g': keyTop,
	'G': keyBottom,
	'v': keyOpen,
	'o': keyUrl,
	'D': keyDownload,
}

var kmap = map[tb.Key]int{
	tb.KeyArrowUp:    keyUp,
	tb.KeyArrowDown:  keyDown,
	tb.KeyArrowLeft:  keyBack,
	tb.KeyBackspace:  keyBack,
	tb.KeyBackspace2: keyBack,
	tb.KeyArrowRight: keyEnter,
	tb.KeyEnter:      keyEnter,
	tb.KeyPgup:       keyPgup,
	tb.KeyPgdn:       keyPgdn,
	tb.KeySpace:      keyPgdn,
	tb.KeyF5:         keyReload,
	tb.KeyHome:       keyTop,
	tb.KeyEnd:        keyBottom,
	tb.KeyCtrlL:      keyRedraw,
}

func getKey(e tb.Event) int {
	if e.Type != tb.EventKey {
		return -1
	}
	if e.Ch == 0 {
		if v, ok := kmap[e.Key]; ok {
			return v
		}
	} else if v, ok := chmap[e.Ch]; ok {
		return v
	}

	return -1
}
