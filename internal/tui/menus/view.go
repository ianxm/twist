package menus

import "twist/internal/log"

func NewViewMenu() Menu {
	return Menu{
		Name:     "View",
		Shortcut: "Alt+V",
		Items: []MenuItem{
			{
				Name:         "Panels",
				IsEnabled:    isConnectedCheck,
				HandleAction: handlePanels,
			},
		},
	}
}

func handlePanels(app AppInterface) error {
	if app.GetPanelsVisible() {
		app.HidePanels()
		log.Info("ViewMenu: Hiding panels")
	} else {
		app.ShowPanels()
		log.Info("ViewMenu: Showing panels")
	}
	return nil
}
