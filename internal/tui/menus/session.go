package menus

func NewSessionMenu() Menu {
	return Menu{
		Name:     "Session",
		Shortcut: "Alt+S",
		Items: []MenuItem{
			{
				Name:         "Connect",
				Shortcut:     "Alt+C",
				CreatesModal: true,
				IsEnabled:    isNotConnectedCheck,
				HandleAction: handleConnect,
			},
			{
				Name:         "Disconnect",
				Shortcut:     "Alt+D",
				IsEnabled:    isConnectedCheck,
				HandleAction: handleDisconnect,
			},
			{
				Name:     "Quit",
				Shortcut: "Alt+Q",
				IsEnabled:    alwaysEnabled,
				HandleAction: handleQuit,
			},
		},
	}
}

func handleConnect(app AppInterface) error {
	app.ShowConnectionDialog()
	return nil
}

func handleDisconnect(app AppInterface) error {
	app.Disconnect()
	return nil
}

func handleQuit(app AppInterface) error {
	app.Exit()
	return nil
}
