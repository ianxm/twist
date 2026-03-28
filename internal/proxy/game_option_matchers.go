package proxy

import "strings"

// GameOptionMatcher parses streaming characters looking for game menu options.
// Each implementation handles a different format (e.g. "<A> Name", "[A] Name", "A. Name").
type GameOptionMatcher interface {
	// ProcessChar feeds a character. Returns (letter, gameName, matched).
	ProcessChar(char rune) (string, string, bool)
	// IsActive returns true if the matcher is mid-parse.
	IsActive() bool
	Reset()
}

// AngleBracketOptionMatcher parses "<X> Game Name" patterns
type AngleBracketOptionMatcher struct {
	state    int
	letter   string
	gameName strings.Builder
}

func (m *AngleBracketOptionMatcher) ProcessChar(char rune) (string, string, bool) {
	switch m.state {
	case 0:
		if char == '<' {
			m.state = 1
		}
	case 1:
		if char >= 'A' && char <= 'Z' {
			m.letter = string(char)
			m.state = 2
		} else {
			m.Reset()
		}
	case 2:
		if char == '>' {
			m.state = 3
		} else {
			m.Reset()
		}
	case 3:
		if char == ' ' || char == '\t' {
			// skip whitespace
		} else if char == '\n' || char == '\r' || char == '[' {
			defer m.Reset()
			return m.letter, strings.TrimSpace(m.gameName.String()), true
		} else {
			m.gameName.WriteRune(char)
			m.state = 4
		}
	case 4:
		if char == '\n' || char == '\r' || char == '[' {
			defer m.Reset()
			return m.letter, strings.TrimSpace(m.gameName.String()), true
		}
		m.gameName.WriteRune(char)
	}
	return "", "", false
}

func (m *AngleBracketOptionMatcher) Reset() {
	m.state = 0
	m.letter = ""
	m.gameName.Reset()
}

func (m *AngleBracketOptionMatcher) IsActive() bool { return m.state != 0 }

// BracketOptionMatcher parses "[X] Game Name" patterns
type BracketOptionMatcher struct {
	state    int
	letter   string
	gameName strings.Builder
}

func (m *BracketOptionMatcher) ProcessChar(char rune) (string, string, bool) {
	switch m.state {
	case 0:
		if char == '[' {
			m.state = 1
		}
	case 1:
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') {
			m.letter = strings.ToUpper(string(char))
			m.state = 2
		} else {
			m.Reset()
		}
	case 2:
		if char == ']' {
			m.state = 3
		} else {
			m.Reset()
		}
	case 3:
		if char == ' ' || char == '\t' {
			// skip whitespace
		} else if char == '\n' || char == '\r' {
			m.Reset()
		} else {
			m.gameName.WriteRune(char)
			m.state = 4
		}
	case 4:
		if char == '\n' || char == '\r' {
			defer m.Reset()
			return m.letter, strings.TrimSpace(m.gameName.String()), true
		}
		m.gameName.WriteRune(char)
	}
	return "", "", false
}

func (m *BracketOptionMatcher) Reset() {
	m.state = 0
	m.letter = ""
	m.gameName.Reset()
}

func (m *BracketOptionMatcher) IsActive() bool { return m.state != 0 }

// DotOptionMatcher parses "X. Game Name" patterns (e.g. "A. Ice9 Pirates Builders")
type DotOptionMatcher struct {
	state    int
	letter   string
	gameName strings.Builder
	prevChar rune
}

func (m *DotOptionMatcher) ProcessChar(char rune) (string, string, bool) {
	switch m.state {
	case 0:
		if ((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')) &&
			(m.prevChar == '\n' || m.prevChar == '\r' ||
				m.prevChar == ' ' || m.prevChar == '\t' || m.prevChar == 0) {
			m.letter = strings.ToUpper(string(char))
			m.state = 1
		}
		m.prevChar = char
	case 1:
		m.prevChar = char
		if char == '.' {
			m.state = 2
		} else {
			m.resetKeepPrev()
		}
	case 2:
		m.prevChar = char
		if char == ' ' || char == '\t' {
			m.state = 3
		} else {
			m.resetKeepPrev()
		}
	case 3:
		m.prevChar = char
		if char == ' ' || char == '\t' {
			// skip extra whitespace
		} else if char == '\n' || char == '\r' {
			m.resetKeepPrev()
		} else {
			m.gameName.WriteRune(char)
			m.state = 4
		}
	case 4:
		m.prevChar = char
		if char == '\n' || char == '\r' {
			letter, name := m.letter, strings.TrimSpace(m.gameName.String())
			m.resetKeepPrev()
			return letter, name, true
		}
		m.gameName.WriteRune(char)
	}
	return "", "", false
}

func (m *DotOptionMatcher) Reset() {
	m.state = 0
	m.letter = ""
	m.gameName.Reset()
	m.prevChar = 0
}

func (m *DotOptionMatcher) IsActive() bool { return m.state != 0 }

func (m *DotOptionMatcher) resetKeepPrev() {
	m.state = 0
	m.letter = ""
	m.gameName.Reset()
}

// DashOptionMatcher parses "X - Game Name" patterns
type DashOptionMatcher struct {
	state    int
	letter   string
	gameName strings.Builder
}

func (m *DashOptionMatcher) ProcessChar(char rune) (string, string, bool) {
	switch m.state {
	case 0:
		if char >= 'A' && char <= 'Z' {
			m.letter = string(char)
			m.state = 1
		}
	case 1:
		if char == ' ' {
			m.state = 2
		} else {
			m.Reset()
		}
	case 2:
		if char == '-' {
			m.state = 3
		} else {
			m.Reset()
		}
	case 3:
		if char == ' ' {
			m.state = 4
		} else {
			m.Reset()
		}
	case 4:
		if char == '\n' || char == '\r' {
			name := strings.TrimSpace(m.gameName.String())
			defer m.Reset()
			if name != "" {
				return m.letter, name, true
			}
		} else {
			m.gameName.WriteRune(char)
		}
	}
	return "", "", false
}

func (m *DashOptionMatcher) Reset() {
	m.state = 0
	m.letter = ""
	m.gameName.Reset()
}

func (m *DashOptionMatcher) IsActive() bool { return m.state != 0 }
