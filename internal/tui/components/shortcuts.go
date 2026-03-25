package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// KeyEventToString converts a tcell.EventKey to a shortcut string like "ctrl+c" or "alt+q"
func KeyEventToString(event *tcell.EventKey) string {
	var parts []string

	if event.Modifiers()&tcell.ModCtrl != 0 {
		parts = append(parts, "ctrl")
	}
	if event.Modifiers()&tcell.ModAlt != 0 {
		parts = append(parts, "alt")
	}
	if event.Modifiers()&tcell.ModShift != 0 {
		parts = append(parts, "shift")
	}

	if event.Rune() != 0 {
		parts = append(parts, string(event.Rune()))
	} else {
		switch event.Key() {
		case tcell.KeyF1:
			parts = append(parts, "f1")
		case tcell.KeyF2:
			parts = append(parts, "f2")
		case tcell.KeyF3:
			parts = append(parts, "f3")
		case tcell.KeyF4:
			parts = append(parts, "f4")
		case tcell.KeyF5:
			parts = append(parts, "f5")
		case tcell.KeyF6:
			parts = append(parts, "f6")
		case tcell.KeyF7:
			parts = append(parts, "f7")
		case tcell.KeyF8:
			parts = append(parts, "f8")
		case tcell.KeyF9:
			parts = append(parts, "f9")
		case tcell.KeyF10:
			parts = append(parts, "f10")
		case tcell.KeyF11:
			parts = append(parts, "f11")
		case tcell.KeyF12:
			parts = append(parts, "f12")
		case tcell.KeyEnter:
			parts = append(parts, "enter")
		case tcell.KeyEscape:
			parts = append(parts, "esc")
		case tcell.KeyTab:
			parts = append(parts, "tab")
		case tcell.KeyBackspace:
			parts = append(parts, "backspace")
		case tcell.KeyDelete:
			parts = append(parts, "delete")
		case tcell.KeyInsert:
			parts = append(parts, "insert")
		case tcell.KeyHome:
			parts = append(parts, "home")
		case tcell.KeyEnd:
			parts = append(parts, "end")
		case tcell.KeyPgUp:
			parts = append(parts, "pageup")
		case tcell.KeyPgDn:
			parts = append(parts, "pagedown")
		case tcell.KeyUp:
			parts = append(parts, "up")
		case tcell.KeyDown:
			parts = append(parts, "down")
		case tcell.KeyLeft:
			parts = append(parts, "left")
		case tcell.KeyRight:
			parts = append(parts, "right")
		default:
			return ""
		}
	}

	return strings.Join(parts, "+")
}
