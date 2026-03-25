package menus

import "twist/internal/log"

func NewTerminalMenu() Menu {
	return Menu{
		Name:     "Terminal",
		Shortcut: "Alt+T",
		Items: []MenuItem{
			{
				Name:         "Clear",
				IsEnabled:    alwaysEnabled,
				HandleAction: handleClear,
			},
		},
	}
}

func handleClear(app AppInterface) error {
	app.ClearTerminal()
	log.Info("TerminalMenu: Cleared terminal")
	return nil
}
