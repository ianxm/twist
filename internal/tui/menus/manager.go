package menus

import (
	"twist/internal/api"
	"twist/internal/log"
)

// AppInterface defines the methods that menu handlers need from TwistApp
type AppInterface interface {
	// Connection management
	Connect(address string)
	Disconnect()
	Exit()
	ShowConnectionDialog()

	// Panel management
	ShowPanels()
	HidePanels()
	GetPanelsVisible() bool

	// Terminal operations
	ClearTerminal()

	// Modal management
	ShowModal(title, text string, buttons []string, callback func(int, string))
	ShowInputDialog(pageName string, dialog interface{})
	ShowBurstDialog(onSend func(string), onCancel func())
	ShowScrollableModal(title, text string, height int)
	CloseModal()

	// Terminal info for dynamic sizing
	GetTerminalWidth() int
	GetTerminalHeight() int

	// Version information
	GetVersion() string
	GetCommit() string
	GetDate() string

	// Proxy API access
	GetProxyAPI() api.ProxyAPI
	IsConnected() bool
}

// MenuManager coordinates menu structure, enablement, and action dispatch.
// It does NOT handle keyboard shortcuts — that's ShortcutManager's job.
type MenuManager struct {
	registry MenuRegistry
	actions  map[string]MenuItem // "MenuName:ItemName" → item
}

func NewMenuManager(registry MenuRegistry) *MenuManager {
	mgr := &MenuManager{
		registry: registry,
		actions:  make(map[string]MenuItem),
	}

	for _, m := range registry.GetMenus() {
		for _, item := range m.Items {
			mgr.actions[m.Name+":"+item.Name] = item
		}
	}
	return mgr
}

func (mm *MenuManager) HandleAction(menuName, itemName string, app AppInterface) error {
	actionKey := menuName + ":" + itemName

	if item, exists := mm.actions[actionKey]; exists {
		if item.IsEnabled != nil && !item.IsEnabled(app) {
			log.Info("MenuManager: Action is disabled", "action", actionKey)
			return nil
		}
		if item.HandleAction != nil {
			return item.HandleAction(app)
		}
	}
	return nil
}

func (mm *MenuManager) GetMenuItems(menuName string) []MenuItem {
	return mm.registry.GetMenuItems(menuName)
}

func (mm *MenuManager) GetEnabledMenuItems(menuName string, app AppInterface) []EnabledMenuItem {
	items := mm.registry.GetMenuItems(menuName)
	if items == nil {
		return []EnabledMenuItem{}
	}

	result := make([]EnabledMenuItem, len(items))
	for i, item := range items {
		enabled := true
		if item.IsEnabled != nil {
			enabled = item.IsEnabled(app)
		}
		result[i] = EnabledMenuItem{MenuItem: item, Enabled: enabled}
	}
	return result
}

// EnabledMenuItem wraps a MenuItem with its enabled status
type EnabledMenuItem struct {
	MenuItem MenuItem
	Enabled  bool
}

func (mm *MenuManager) GetMenuOptions(menuName string) []string {
	return mm.registry.GetMenuOptions(menuName)
}

func (mm *MenuManager) GetMenuNames() []string {
	return mm.registry.GetMenuNames()
}

func (mm *MenuManager) GetDropdownPosition(menuName string) int {
	return mm.registry.CalculateDropdownPosition(menuName)
}

func (mm *MenuManager) ActionCreatesModal(menuName, action string) bool {
	items := mm.GetMenuItems(menuName)
	for _, item := range items {
		if item.Name == action {
			return item.CreatesModal
		}
	}
	return false
}

// GetRegistry returns the underlying registry for shortcut registration
func (mm *MenuManager) GetRegistry() MenuRegistry {
	return mm.registry
}
