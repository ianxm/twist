package handlers

import (
	"twist/internal/log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InputMode represents the current input mode
type InputMode int

const (
	InputModeMenu InputMode = iota
	InputModeTerminal
	InputModeModal
)

// InputHandler manages input handling for the application
type InputHandler struct {
	app               *tview.Application
	inputMode         InputMode
	modalVisible      bool
	isDropdownVisible func() bool

	// Callbacks
	onConnect     func(string)
	onDisconnect  func()
	onExit        func()
	onShowModal   func(string, []string, func(string))
	onCloseModal  func()
	onSendCommand func(string)
}

// NewInputHandler creates a new input handler
func NewInputHandler(app *tview.Application) *InputHandler {
	return &InputHandler{
		app:       app,
		inputMode: InputModeMenu,
	}
}

// SetCallbacks sets the callback functions
func (ih *InputHandler) SetCallbacks(
	onConnect func(string),
	onDisconnect func(),
	onExit func(),
	onShowModal func(string, []string, func(string)),
	onCloseModal func(),
	onSendCommand func(string),
) {
	ih.onConnect = onConnect
	ih.onDisconnect = onDisconnect
	ih.onExit = onExit
	ih.onShowModal = onShowModal
	ih.onCloseModal = onCloseModal
	ih.onSendCommand = onSendCommand
}

// SetDropdownVisibilityChecker sets the function to check if dropdown is visible
func (ih *InputHandler) SetDropdownVisibilityChecker(isDropdownVisible func() bool) {
	ih.isDropdownVisible = isDropdownVisible
}

// SetInputMode sets the current input mode
func (ih *InputHandler) SetInputMode(mode InputMode) {
	ih.inputMode = mode
}

// SetModalVisible sets the modal visibility state
func (ih *InputHandler) SetModalVisible(visible bool) {
	ih.modalVisible = visible
}

// HandleKeyEvent handles key events based on current input mode.
// Note: Menu shortcuts (Alt+key) are handled by handleGlobalKeys before this is called.
func (ih *InputHandler) HandleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	if ih.modalVisible {
		return ih.handleModalInput(event)
	}

	switch ih.inputMode {
	case InputModeMenu:
		return ih.handleMenuInput(event)
	case InputModeTerminal:
		return ih.handleTerminalInput(event)
	}

	return event
}

// handleMenuInput handles input in menu navigation mode
func (ih *InputHandler) handleMenuInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		ih.SetInputMode(InputModeTerminal)
		return nil
	}
	return event
}

// handleTerminalInput handles input in terminal mode
func (ih *InputHandler) handleTerminalInput(event *tcell.EventKey) *tcell.EventKey {
	// Don't send Control key combinations (let tview handle them)
	if event.Modifiers()&tcell.ModCtrl != 0 {
		return event
	}
	// Don't send Alt key combinations (handled by handleGlobalKeys via MenuManager)
	if event.Modifiers()&tcell.ModAlt != 0 {
		return event
	}

	switch event.Key() {
	case tcell.KeyTab:
		ih.SetInputMode(InputModeMenu)
		return nil
	case tcell.KeyEscape:
		ih.SetInputMode(InputModeMenu)
		return nil
	}

	// Let the terminal component handle all other keys directly
	return event
}

// handleModalInput handles input when a modal dialog is visible
func (ih *InputHandler) handleModalInput(event *tcell.EventKey) *tcell.EventKey {
	log.Info("handleModalInput", "key", event.Key(), "rune", event.Rune())

	switch event.Key() {
	case tcell.KeyEscape:
		if ih.onCloseModal != nil {
			ih.onCloseModal()
		}
		return nil
	}

	// Pass all other events to the focused component (form, modal, etc.)
	return event
}
