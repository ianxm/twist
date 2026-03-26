package menus

// MenuRegistry provides centralized menu configuration
type MenuRegistry interface {
	GetMenus() []Menu
	GetMenuNames() []string
	GetMenuItems(menuName string) []MenuItem
	GetMenuOptions(menuName string) []string
	CalculateDropdownPosition(menuName string) int
}

type AppMenuRegistry struct {
	menus []Menu // Ordered slice preserves insertion order
}

func NewAppMenuRegistry() *AppMenuRegistry {
	mr := &AppMenuRegistry{}
	mr.register(NewSessionMenu())
	mr.register(NewViewMenu())
	mr.register(NewScriptsMenu())
	mr.register(NewTerminalMenu())
	mr.register(NewHelpMenu())
	return mr
}

func (mr *AppMenuRegistry) register(menu Menu) {
	mr.menus = append(mr.menus, menu)
}

func (mr *AppMenuRegistry) GetMenuNames() []string {
	names := make([]string, len(mr.menus))
	for i, menu := range mr.menus {
		names[i] = menu.Name
	}
	return names
}

func (mr *AppMenuRegistry) GetMenuItems(menuName string) []MenuItem {
	for _, menu := range mr.menus {
		if menu.Name == menuName {
			return menu.Items
		}
	}
	return nil
}

func (mr *AppMenuRegistry) GetMenuOptions(menuName string) []string {
	items := mr.GetMenuItems(menuName)
	options := make([]string, len(items))
	for i, item := range items {
		options[i] = item.Name
	}
	return options
}

func (mr *AppMenuRegistry) GetMenu(menuName string) (Menu, bool) {
	for _, menu := range mr.menus {
		if menu.Name == menuName {
			return menu, true
		}
	}
	return Menu{}, false
}

func (mr *AppMenuRegistry) GetMenus() []Menu {
	return mr.menus
}

// CalculateDropdownPosition calculates the X position for a dropdown menu
func (mr *AppMenuRegistry) CalculateDropdownPosition(menuName string) int {
	position := 1 // Start with 1 space padding
	for _, menu := range mr.menus {
		if menu.Name == menuName {
			return position
		}
		position += len(menu.Name) + 2
	}
	return 1
}
