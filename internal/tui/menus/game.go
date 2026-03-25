package menus

func NewGameMenu() Menu {
	return Menu{
		Name:     "Game",
		Shortcut: "Alt+G",
		Items: []MenuItem{
			{
				Name:         "Bubbles",
				CreatesModal: true,
				IsEnabled:    isConnectedCheck,
				HandleAction: handleBubbles,
			},
			{
				Name:         "Paired Ports",
				CreatesModal: true,
				IsEnabled:    isConnectedCheck,
				HandleAction: handlePairedPorts,
			},
		},
	}
}

func handleBubbles(app AppInterface) error {
	app.ShowModal("Bubbles", "Bubbles\n\nbubble 1", []string{"Close"},
		func(buttonIndex int, buttonLabel string) {
			app.CloseModal()
		})
	return nil
}

func handlePairedPorts(app AppInterface) error {
	app.ShowModal("Paired Ports", "Paired Ports\n\npair 1", []string{"Close"},
		func(buttonIndex int, buttonLabel string) {
			app.CloseModal()
		})
	return nil
}
