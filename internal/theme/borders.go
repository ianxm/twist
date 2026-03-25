package theme

import "github.com/rivo/tview"

// MenuBorderStyle defines predefined border character sets for menus
type MenuBorderStyle int

const (
	MenuBorderStyleSingle  MenuBorderStyle = iota
	MenuBorderStyleDouble
	MenuBorderStyleHeavy
	MenuBorderStyleRounded
)

// BorderChars defines the characters used for drawing borders
type BorderChars struct {
	Normal MenuBorderStyle
	Focus  MenuBorderStyle
}

func NewBorderChars(normal, focus MenuBorderStyle) *BorderChars {
	return &BorderChars{Normal: normal, Focus: focus}
}

func NewSimpleBorderChars(style MenuBorderStyle) *BorderChars {
	return NewBorderChars(style, style)
}

func getBorderRunes(style MenuBorderStyle) (horizontal, vertical, topLeft, topRight, bottomLeft, bottomRight rune) {
	switch style {
	case MenuBorderStyleDouble:
		return '═', '║', '╔', '╗', '╚', '╝'
	case MenuBorderStyleHeavy:
		return '━', '┃', '┏', '┓', '┗', '┛'
	case MenuBorderStyleRounded:
		return '─', '│', '╭', '╮', '╰', '╯'
	default: // Single
		return '─', '│', '┌', '┐', '└', '┘'
	}
}

// ToTviewBorders converts BorderChars to tview's global Borders struct type
func (bc *BorderChars) ToTviewBorders() struct {
	Horizontal       rune
	Vertical         rune
	TopLeft          rune
	TopRight         rune
	BottomLeft       rune
	BottomRight      rune
	LeftT            rune
	RightT           rune
	TopT             rune
	BottomT          rune
	Cross            rune
	HorizontalFocus  rune
	VerticalFocus    rune
	TopLeftFocus     rune
	TopRightFocus    rune
	BottomLeftFocus  rune
	BottomRightFocus rune
} {
	normalH, normalV, normalTL, normalTR, normalBL, normalBR := getBorderRunes(bc.Normal)
	focusH, focusV, focusTL, focusTR, focusBL, focusBR := getBorderRunes(bc.Focus)

	return struct {
		Horizontal       rune
		Vertical         rune
		TopLeft          rune
		TopRight         rune
		BottomLeft       rune
		BottomRight      rune
		LeftT            rune
		RightT           rune
		TopT             rune
		BottomT          rune
		Cross            rune
		HorizontalFocus  rune
		VerticalFocus    rune
		TopLeftFocus     rune
		TopRightFocus    rune
		BottomLeftFocus  rune
		BottomRightFocus rune
	}{
		Horizontal:       normalH,
		Vertical:         normalV,
		TopLeft:          normalTL,
		TopRight:         normalTR,
		BottomLeft:       normalBL,
		BottomRight:      normalBR,
		LeftT:            '├',
		RightT:           '┤',
		TopT:             '┬',
		BottomT:          '┴',
		Cross:            '┼',
		HorizontalFocus:  focusH,
		VerticalFocus:    focusV,
		TopLeftFocus:     focusTL,
		TopRightFocus:    focusTR,
		BottomLeftFocus:  focusBL,
		BottomRightFocus: focusBR,
	}
}

// ApplyBordersForDraw temporarily applies custom borders, calls draw, then restores.
// This is the standard pattern for per-component border customization in tview.
func (bc *BorderChars) ApplyBordersForDraw(draw func()) {
	originalBorders := tview.Borders
	tview.Borders = bc.ToTviewBorders()
	draw()
	tview.Borders = originalBorders
}
