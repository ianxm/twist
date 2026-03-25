package components

import (
	"strings"
	"twist/internal/theme"
	"twist/internal/tui/menus"

	"github.com/rivo/tview"
)

// MenuComponent manages the menu bar component
type MenuComponent struct {
	view          *tview.TextView
	dropdown      *DropdownMenu
	dropdownPages *tview.Pages
	activeMenu    string
	connected     bool
	targetWidth   int // Target width to match status bar

	menuManager MenuManagerInterface
}

// MenuManagerInterface defines what the menu component needs from the menu manager
type MenuManagerInterface interface {
	GetMenuNames() []string
	GetDropdownPosition(menuName string) int
}

// NewMenuComponent creates a new menu component
func NewMenuComponent(menuManager MenuManagerInterface) *MenuComponent {
	menuBar := theme.NewMenuBar().
		SetDynamicColors(true).
		SetRegions(false).
		SetWrap(false).
		SetTextAlign(tview.AlignLeft)

	// Set background to theme background to prevent color bleeding
	currentTheme := theme.Current()
	defaultColors := currentTheme.DefaultColors()
	menuBar.SetBackgroundColor(defaultColors.Background)

	mc := &MenuComponent{
		view:          menuBar,
		dropdown:      NewDropdownMenu(),
		dropdownPages: tview.NewPages(),
		activeMenu:    "",
		connected:     false,
		targetWidth:   0,
		menuManager:   menuManager,
	}

	mc.updateMenuWithHighlight()

	return mc
}

func (mc *MenuComponent) GetView() *tview.TextView {
	return mc.view
}

func (mc *MenuComponent) UpdateMenu(connected bool) {
	mc.connected = connected
	mc.updateMenuWithHighlight()
}

func (mc *MenuComponent) updateMenuWithHighlight() {
	currentTheme := theme.Current()
	menuColors := currentTheme.MenuColors()

	menuNames := mc.menuManager.GetMenuNames()

	menuText := " "
	for _, menu := range menuNames {
		if menu == mc.activeMenu {
			menuText += "[white:red]" + menu + "[:-]  "
		} else {
			menuText += menu + "  "
		}
	}
	menuText += " "

	plainText := mc.stripColorTags(menuText)
	contentLength := len(plainText)

	minPanelWidth := 110
	targetWidth := mc.targetWidth
	if targetWidth == 0 {
		targetWidth = minPanelWidth
	}

	if targetWidth > contentLength {
		paddingNeeded := targetWidth - contentLength
		menuText += strings.Repeat(" ", paddingNeeded)
	}

	finalText := "[:" + menuColors.Background.String() + "]" + menuText + "[-:-]"
	mc.view.SetText(finalText)
}

func (mc *MenuComponent) SetConnectedMenu() {
	mc.UpdateMenu(true)
}

func (mc *MenuComponent) SetDisconnectedMenu() {
	mc.UpdateMenu(false)
}

// ShowDropdown displays a dropdown menu
func (mc *MenuComponent) ShowDropdown(menuName string, items []menus.MenuItem, callback func(string), navCallback func(string)) *tview.Flex {
	leftOffset := mc.menuManager.GetDropdownPosition(menuName)

	mc.dropdown = NewDropdownMenu()
	mc.dropdown.SetNavigationCallback(navCallback)

	return mc.dropdown.Show(menuName, items, leftOffset, callback)
}

func (mc *MenuComponent) HideDropdown() {
	if mc.dropdown != nil {
		mc.dropdown.Hide()
	}
}

func (mc *MenuComponent) IsDropdownVisible() bool {
	if mc.dropdown != nil {
		return mc.dropdown.IsVisible()
	}
	return false
}

func (mc *MenuComponent) GetDropdownList() *TwistMenu {
	if mc.dropdown != nil {
		return mc.dropdown.GetList()
	}
	return nil
}

func (mc *MenuComponent) GetCurrentDropdown() *DropdownMenu {
	return mc.dropdown
}

func (mc *MenuComponent) SetTargetWidth(width int) {
	mc.targetWidth = width
	mc.updateMenuWithHighlight()
}

// stripColorTags removes tview color tags from text to calculate actual display length
func (mc *MenuComponent) stripColorTags(text string) string {
	result := text
	result = strings.ReplaceAll(result, "[-]", "")
	result = strings.ReplaceAll(result, "[-:-]", "")

	for strings.Contains(result, "[") && strings.Contains(result, "]") {
		start := strings.Index(result, "[")
		end := strings.Index(result[start:], "]")
		if end != -1 {
			result = result[:start] + result[start+end+1:]
		} else {
			break
		}
	}

	return result
}
