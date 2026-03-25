package menus

func NewHelpMenu() Menu {
	return Menu{
		Name:     "Help",
		Shortcut: "Alt+H",
		Items: []MenuItem{
			{
				Name:         "Keyboard Shortcuts",
				Shortcut:     "F1",
				CreatesModal: true,
				IsEnabled:    alwaysEnabled,
				HandleAction: handleKeyboardShortcuts,
			},
			{
				Name:         "About",
				CreatesModal: true,
				IsEnabled:    alwaysEnabled,
				HandleAction: handleAbout,
			},
			{
				Name:         "User Manual",
				CreatesModal: true,
				IsEnabled:    alwaysEnabled,
				HandleAction: handleUserManual,
			},
		},
	}
}

func handleKeyboardShortcuts(app AppInterface) error {
	helpText := "TWIST Terminal Interface - Keyboard Shortcuts\n\n" +
		"Menu Navigation:\n" +
		"Alt+S = Session menu\n" +
		"Alt+V = View menu\n" +
		"Alt+T = Terminal menu\n" +
		"Alt+H = Help menu\n\n" +
		"Quick Actions:\n" +
		"Alt+C = Connect to server\n" +
		"Alt+D = Disconnect from server\n" +
		"Alt+Q = Quit application\n\n" +
		"Function Keys:\n" +
		"F1 = Show this help screen\n" +
		"ESC = Close dialogs and menus\n\n" +
		"Global Keys:\n" +
		"Ctrl+C = Exit application"

	app.ShowModal("Keyboard Shortcuts", helpText, []string{"Close"},
		func(buttonIndex int, buttonLabel string) {
			app.CloseModal()
		})
	return nil
}

func handleAbout(app AppInterface) error {
	aboutText := "TWIST Terminal Interface\n\n" +
		"Version: " + app.GetVersion() + "\n" +
		"Commit: " + app.GetCommit() + "\n" +
		"Build Date: " + app.GetDate() + "\n\n" +
		"A Trade Wars 2002 proxy client with scripting support.\n" +
		"Built with Go and tview for cross-platform terminal interfaces.\n\n" +
		"Features:\n" +
		"• Real-time game data parsing\n" +
		"• TWX-compatible scripting engine\n" +
		"• Sector mapping and navigation\n" +
		"• Database integration\n" +
		"• Cross-platform terminal UI"

	app.ShowModal("About TWIST", aboutText, []string{"Close"},
		func(buttonIndex int, buttonLabel string) {
			app.CloseModal()
		})
	return nil
}

func handleUserManual(app AppInterface) error {
	app.ShowModal("User Manual",
		"User Manual feature not yet implemented.\n\n"+
			"For now, use F1 or Help → Keyboard Shortcuts\n"+
			"for basic usage information.",
		[]string{"OK"},
		func(buttonIndex int, buttonLabel string) {
			app.CloseModal()
		})
	return nil
}
