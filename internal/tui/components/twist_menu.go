package components

import (
	"twist/internal/theme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TwistMenu wraps a tview.List to provide custom border character support.
// It accepts plain strings for labels/shortcuts — no dependency on menu domain types.
type TwistMenu struct {
	*tview.List
	borderChars *theme.BorderChars
	callbacks   []func()
}

// NewTwistMenu creates a new TwistMenu with optional custom border characters
func NewTwistMenu(borderChars *theme.BorderChars) *TwistMenu {
	tm := &TwistMenu{
		List:        tview.NewList(),
		borderChars: borderChars,
		callbacks:   make([]func(), 0),
	}

	tm.List.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			currentItem := tm.List.GetCurrentItem()
			if currentItem >= 0 && currentItem < len(tm.callbacks) && tm.callbacks[currentItem] != nil {
				tm.callbacks[currentItem]()
			}
			return nil
		}
		return event
	})

	return tm
}

// NewThemedTwistMenu creates a TwistMenu with current theme styling applied.
// Dropdown.Show() creates a fresh one each time to ensure proper sizing per menu.
func NewThemedTwistMenu() *TwistMenu {
	currentTheme := theme.Current()
	colors := currentTheme.MenuColors()
	borderStyle := currentTheme.MenuBorderStyle()
	borderChars := theme.NewSimpleBorderChars(borderStyle)

	menu := NewTwistMenu(borderChars)

	menu.SetBackgroundColor(colors.Background)
	mainStyle := tcell.StyleDefault.
		Foreground(colors.Foreground).
		Background(colors.Background)
	menu.SetMainTextStyle(mainStyle)
	menu.SetSelectedTextColor(colors.SelectedFg)
	menu.SetSelectedBackgroundColor(colors.SelectedBg)
	menu.SetBorderColor(colors.Foreground)
	menu.SetBorder(true)

	return menu
}

func (tm *TwistMenu) SetBorderChars(borderChars *theme.BorderChars) *TwistMenu {
	tm.borderChars = borderChars
	return tm
}

// SetMenuItems replaces all menu items. Accepts plain string slices for labels and shortcuts.
func (tm *TwistMenu) SetMenuItems(labels []string, shortcuts []string, callbacks []func()) {
	tm.List.Clear()
	tm.callbacks = callbacks

	hasAnyShortcuts := false
	maxLabelWidth := 0
	maxShortcutWidth := 0

	for i, label := range labels {
		if len(label) > maxLabelWidth {
			maxLabelWidth = len(label)
		}
		if i < len(shortcuts) && shortcuts[i] != "" {
			hasAnyShortcuts = true
			if len(shortcuts[i]) > maxShortcutWidth {
				maxShortcutWidth = len(shortcuts[i])
			}
		}
	}

	for i, label := range labels {
		shortcut := ""
		if i < len(shortcuts) {
			shortcut = shortcuts[i]
		}

		displayText := label
		if hasAnyShortcuts && shortcut != "" {
			padding := maxLabelWidth - len(label) + 3
			spaces := ""
			for j := 0; j < padding; j++ {
				spaces += " "
			}
			displayText = label + spaces + shortcut
		}

		tm.List.AddItem(displayText, "", 0, nil)
	}
}

// GetCallbacks returns the current menu item callbacks
func (tm *TwistMenu) GetCallbacks() []func() {
	return tm.callbacks
}

// Draw overrides the default tview.List Draw to apply custom border characters.
func (tm *TwistMenu) Draw(screen tcell.Screen) {
	if tm.borderChars != nil {
		tm.borderChars.ApplyBordersForDraw(func() {
			tm.List.Draw(screen)
		})
	} else {
		tm.List.Draw(screen)
	}
}
