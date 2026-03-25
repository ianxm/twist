package components

import (
	"twist/internal/theme"
	"twist/internal/tui/menus"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// DropdownMenu represents a dropdown menu component
type DropdownMenu struct {
	list               *TwistMenu
	wrapper            *tview.Flex
	visible            bool
	callback           func(string)
	navigationCallback func(direction string) // "left" or "right"
}

// NewDropdownMenu creates a new dropdown menu
func NewDropdownMenu() *DropdownMenu {
	list := NewThemedTwistMenu()

	dm := &DropdownMenu{
		list:    list,
		visible: false,
	}

	list.SetInputCapture(dm.handleInput)

	return dm
}

// Show displays the dropdown menu
func (dm *DropdownMenu) Show(menuName string, items []menus.MenuItem, leftOffset int, callback func(string)) *tview.Flex {
	dm.callback = callback
	dm.visible = true

	dm.list = NewThemedTwistMenu()
	dm.list.SetInputCapture(dm.handleInput)

	// Build string slices from menu items
	labels := make([]string, len(items))
	shortcuts := make([]string, len(items))
	callbacks := make([]func(), len(items))

	for i, item := range items {
		label := item.Name
		labels[i] = label
		shortcuts[i] = item.Shortcut
		callbacks[i] = func() {
			if dm.callback != nil {
				dm.callback(label)
			}
			dm.Hide()
		}
	}

	dm.list.SetMenuItems(labels, shortcuts, callbacks)

	// Calculate dropdown width based on content
	maxLabelWidth := 0
	maxShortcutWidth := 0
	hasAnyShortcuts := false

	for _, item := range items {
		if len(item.Name) > maxLabelWidth {
			maxLabelWidth = len(item.Name)
		}
		if item.Shortcut != "" {
			hasAnyShortcuts = true
			if len(item.Shortcut) > maxShortcutWidth {
				maxShortcutWidth = len(item.Shortcut)
			}
		}
	}

	var dropdownWidth int
	if hasAnyShortcuts {
		// Width = longest label + gap (3) + longest shortcut + border padding (6)
		dropdownWidth = maxLabelWidth + 3 + maxShortcutWidth + 6
	} else {
		// Width = longest label + border padding
		dropdownWidth = maxLabelWidth + 6
	}
	if dropdownWidth < 15 {
		dropdownWidth = 15 // Minimum width
	}

	itemCount := len(items)
	calculatedHeight := calculateMenuHeight(itemCount)

	// Apply theme background to spacers
	currentTheme := theme.Current()
	defaultColors := currentTheme.DefaultColors()
	_ = defaultColors

	dm.wrapper = tview.NewFlex().
		AddItem(nil, leftOffset, 0, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 1, 0, false).                       // Top spacing (1 row below menu bar)
			AddItem(dm.list, calculatedHeight, 0, true).     // Menu list
			AddItem(nil, 0, 1, false), dropdownWidth, 0, true).
		AddItem(nil, 0, 1, false)

	return dm.wrapper
}

// Hide hides the dropdown menu
func (dm *DropdownMenu) Hide() {
	dm.visible = false
	dm.callback = nil
}

// IsVisible returns whether the dropdown is currently visible
func (dm *DropdownMenu) IsVisible() bool {
	return dm.visible
}

// GetList returns the underlying list component for focus management
func (dm *DropdownMenu) GetList() *TwistMenu {
	return dm.list
}

// SetNavigationCallback sets the callback for left/right arrow navigation
func (dm *DropdownMenu) SetNavigationCallback(callback func(direction string)) {
	dm.navigationCallback = callback
}

// SetItemEnabled sets the enabled/disabled state of a specific menu item by index
// SetItemEnabled sets the enabled/disabled state of a specific menu item by index
func (dm *DropdownMenu) SetItemEnabled(itemIndex int, enabled bool) {
	if dm.list == nil {
		return
	}

	itemCount := dm.list.GetItemCount()
	if itemIndex < 0 || itemIndex >= itemCount {
		return
	}

	mainText, _ := dm.list.GetItemText(itemIndex)
	cleanText := dm.stripColorTags(mainText)

	var styledText string
	if enabled {
		styledText = cleanText
	} else {
		styledText = "[darkgray]" + cleanText + "[white]" // Grayed out
	}

	dm.list.SetItemText(itemIndex, styledText, "")
}

// stripColorTags removes tview color tags from text
func (dm *DropdownMenu) stripColorTags(text string) string {
	result := ""
	inTag := false
	for _, char := range text {
		if char == '[' {
			inTag = true
		} else if char == ']' && inTag {
			inTag = false
		} else if !inTag {
			result += string(char)
		}
	}
	return result
}

func (dm *DropdownMenu) handleInput(event *tcell.EventKey) *tcell.EventKey {
	// Handle menu navigation (left/right between different menus)
	switch event.Key() {
	case tcell.KeyLeft:
		if dm.navigationCallback != nil {
			dm.navigationCallback("left")
		}
		return nil
	case tcell.KeyRight:
		if dm.navigationCallback != nil {
			dm.navigationCallback("right")
		}
		return nil
	case tcell.KeyEnter:
		// Trigger menu item and consume to prevent terminal input
		currentItem := dm.list.GetCurrentItem()
		if currentItem >= 0 && currentItem < dm.list.GetItemCount() {
			callbacks := dm.list.GetCallbacks()
			if currentItem < len(callbacks) && callbacks[currentItem] != nil {
				callbacks[currentItem]()
			}
		}
		return nil
	default:
		return event
	}
}

// calculateMenuHeight determines the proper height for a dropdown menu.
// Formula derived from empirical data: 1 item → 4 rows, 2→6, 3→8.
// Pattern: height = 2 * itemCount + 2
func calculateMenuHeight(itemCount int) int {
	if itemCount <= 0 {
		return 4
	}
	return 2*itemCount + 2
}
