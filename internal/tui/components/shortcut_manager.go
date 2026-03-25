package components

import "strings"

// ShortcutManager handles app-wide keyboard shortcut dispatch.
// Menus register their shortcuts here at startup; handleGlobalKeys() calls Handle().
type ShortcutManager struct {
	shortcuts map[string]func() // normalized shortcut string → callback
}

func NewShortcutManager() *ShortcutManager {
	return &ShortcutManager{
		shortcuts: make(map[string]func()),
	}
}

// Register adds a shortcut → callback mapping. Shortcut strings are normalized
// so "Alt-Q", "alt+q", "Alt+Q" all match.
func (sm *ShortcutManager) Register(shortcut string, callback func()) {
	if shortcut != "" && callback != nil {
		sm.shortcuts[normalize(shortcut)] = callback
	}
}

// Handle checks if a key string matches a registered shortcut and fires it.
// Returns true if the shortcut was handled.
func (sm *ShortcutManager) Handle(key string) bool {
	if callback, exists := sm.shortcuts[normalize(key)]; exists {
		callback()
		return true
	}
	return false
}

// GetAll returns all registered shortcuts as shortcut→description pairs.
// Useful for generating help screens.
func (sm *ShortcutManager) GetAll() map[string]string {
	result := make(map[string]string, len(sm.shortcuts))
	for k := range sm.shortcuts {
		result[k] = "" // descriptions can be added later if needed
	}
	return result
}

func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", "+")
	return s
}
