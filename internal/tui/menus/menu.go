package menus

// MenuItem represents a single menu item with optional shortcut and action handler
type MenuItem struct {
	Name         string
	Shortcut     string
	CreatesModal bool
	IsEnabled    func(app AppInterface) bool
	HandleAction func(app AppInterface) error
}

// Menu defines a complete menu with all its properties
type Menu struct {
	Name     string
	Shortcut string
	Items    []MenuItem
}

// Helper functions for menu item enablement checks
func isConnectedCheck(app AppInterface) bool {
	proxyAPI := app.GetProxyAPI()
	return proxyAPI != nil && proxyAPI.IsConnected()
}

func isNotConnectedCheck(app AppInterface) bool {
	proxyAPI := app.GetProxyAPI()
	return proxyAPI == nil || !proxyAPI.IsConnected()
}

func alwaysEnabled(app AppInterface) bool {
	return true
}
