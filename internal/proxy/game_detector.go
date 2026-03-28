package proxy

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"twist/internal/ansi"
	"twist/internal/log"
	"twist/internal/proxy/database"
	"twist/internal/proxy/scripting"
)

// Token types for game detection
type TokenType int

const (
	TokenError TokenType = iota
	TokenEOF
	TokenText           // Regular text to ignore
	TokenGameMenu       // "Select a game :" pattern
	TokenGameOption     // "<A> Game Name" pattern
	TokenIsolatedLetter // Single letter game selection
	TokenGameStart      // "Show today's log? (Y/N)" pattern
	TokenGameExit       // Exit patterns
	TokenMainMenu       // Return to main menu patterns
	TokenUserPrompt     // User input prompt patterns like "Your choice: " or "Enter selection: "
)

type GameDetectionState int

const (
	StateIdle GameDetectionState = iota
	StateGameMenuVisible
	StateGameSelected
	StateGameActive // Combines starting and active - game is running
)

// String returns a string representation of the GameDetectionState
func (s GameDetectionState) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateGameMenuVisible:
		return "GameMenuVisible"
	case StateGameSelected:
		return "GameSelected"
	case StateGameActive:
		return "GameActive"
	default:
		return "Unknown"
	}
}

// ConnectionInfo holds server connection details
type ConnectionInfo struct {
	Host string
	Port string
}

// Token represents a lexed token
type Token struct {
	Type     TokenType
	Value    string
	Letter   string // For game options and selections
	GameName string // For game options
	Pos      int    // Position in input
}

// PatternMatcher tracks progress matching a specific pattern
type PatternMatcher struct {
	pattern   string               // Pattern to match
	position  int                  // Current position in pattern
	tokenType TokenType            // Token to emit on match
	buffer    string               // Buffer for this pattern
	isActive  bool                 // Whether this matcher is currently active
	states    map[GameDetectionState]bool // Which states this pattern is checked in (nil = all states)
}

// stateFn represents the state of the scanner as a function that returns the next state
type stateFn func(*GameDetector) stateFn

// gameDetectorState represents the immutable state that can be atomically swapped
type gameDetectorState struct {
	currentState       GameDetectionState
	selectedGame       string
	selectedLetter     string
	gameOptions        map[string]string
	expectingUserInput bool
}

// GameDetector is a streaming lexer for game detection
type GameDetector struct {
	mu sync.RWMutex

	// Connection info
	serverHost string
	serverPort string

	// Streaming input handling
	currentBuffer   string                     // Small buffer for current potential match
	patternMatchers map[string]*PatternMatcher // Active pattern matchers
	recentContent   string                     // Larger buffer for context analysis (last ~500 chars)

	// ANSI stripping for streaming content
	ansiStripper *ansi.StreamingStripper

	// Detection state - atomic pointer to immutable state
	state atomic.Pointer[gameDetectorState]

	// Token channel
	tokens chan Token

	// Instance-specific state machines
	optionMatchers []GameOptionMatcher
	iLetterState   *isolatedLetterState

	// Database management
	currentDatabase      database.Database
	currentScriptManager *scripting.ScriptManager

	// Callbacks
	onDatabaseLoaded       func(db database.Database, scriptManager *scripting.ScriptManager) error
	onDatabaseStateChanged func(gameName, serverHost, serverPort, dbName string, isLoaded bool)

	// Timing
	lastActivity     time.Time
	detectionTimeout time.Duration
}

// NewGameDetector creates a new lexer-based game detector
func NewGameDetector(connInfo ConnectionInfo) *GameDetector {
	l := &GameDetector{
		serverHost:       connInfo.Host,
		serverPort:       connInfo.Port,
		tokens:           make(chan Token, 100), // Buffered channel
		detectionTimeout: time.Minute * 5,
		patternMatchers:  make(map[string]*PatternMatcher),
		ansiStripper:     ansi.NewStreamingStripper(),
		// Initialize instance-specific state machines
		optionMatchers: []GameOptionMatcher{
			&AngleBracketOptionMatcher{},
			&BracketOptionMatcher{},
			&DotOptionMatcher{},
		},
		iLetterState: &isolatedLetterState{},
	}

	// Initialize atomic state
	initialState := &gameDetectorState{
		currentState:       StateIdle,
		selectedGame:       "",
		gameOptions:        make(map[string]string),
		expectingUserInput: false,
	}
	l.state.Store(initialState)

	// Initialize pattern matchers
	l.initializePatterns()

	return l
}

// updateState atomically updates the game detector state
func (l *GameDetector) updateState(updateFn func(*gameDetectorState) *gameDetectorState) {
	for {
		oldState := l.state.Load()
		newState := updateFn(oldState)
		if l.state.CompareAndSwap(oldState, newState) {
			break
		}
		// Retry if CAS failed due to concurrent update
	}
}

// copyState creates a copy of the current state for modification
func copyState(s *gameDetectorState) *gameDetectorState {
	if s == nil {
		return &gameDetectorState{
			currentState:       StateIdle,
			selectedGame:       "",
			gameOptions:        make(map[string]string),
			expectingUserInput: false,
		}
	}

	// Deep copy the map
	gameOptionsCopy := make(map[string]string, len(s.gameOptions))
	for k, v := range s.gameOptions {
		gameOptionsCopy[k] = v
	}

	return &gameDetectorState{
		currentState:       s.currentState,
		selectedGame:       s.selectedGame,
		selectedLetter:     s.selectedLetter,
		gameOptions:        gameOptionsCopy,
		expectingUserInput: s.expectingUserInput,
	}
}

// initializePatterns sets up all the pattern matchers
func (l *GameDetector) initializePatterns() {
	allStates := map[GameDetectionState]bool{
		StateIdle: true, StateGameMenuVisible: true,
		StateGameSelected: true, StateGameActive: true,
	}
	menuOnly := map[GameDetectionState]bool{
		StateGameMenuVisible: true,
	}
	menuAndIdle := map[GameDetectionState]bool{
		StateIdle: true, StateGameMenuVisible: true,
	}
	waitingForGame := map[GameDetectionState]bool{
		StateGameMenuVisible: true, StateGameSelected: true,
	}

	type patternDef struct {
		tokenType TokenType
		states    map[GameDetectionState]bool
	}

	patterns := map[string]patternDef{
		// Exit patterns — any state
		"Goodbye":                     {TokenGameExit, allStates},
		"Thank you for playing":       {TokenGameExit, allStates},
		"Connection terminated":       {TokenGameExit, allStates},
		"Disconnected":                {TokenGameExit, allStates},
		"session has been terminated": {TokenGameExit, allStates},
		"CRITICAL INACTIVITY:":        {TokenGameExit, allStates},
		"...Now leaving Trade Wars":   {TokenGameExit, allStates},

		// Main menu detection — any state
		"TWGS v":                {TokenMainMenu, allStates},
		"TradeWars Game Server": {TokenMainMenu, allStates},

		// Game menu pattern — idle/menu
		"Select a game :": {TokenGameMenu, menuAndIdle},

		// Game start patterns — menu visible or game selected
		"Show today's log?": {TokenGameStart, waitingForGame},
		"Command [TL=":      {TokenGameStart, waitingForGame},

		// User prompt patterns — menu visible only
		"Your choice: ":               {TokenUserPrompt, menuOnly},
		"Enter selection: ":           {TokenUserPrompt, menuOnly},
		"Choice: ":                    {TokenUserPrompt, menuOnly},
		"Enter your choice: ":         {TokenUserPrompt, menuOnly},
		"Please enter your choice: ":  {TokenUserPrompt, menuOnly},
		"Selection: ":                 {TokenUserPrompt, menuOnly},
		"Your Command:":               {TokenUserPrompt, menuOnly},
		"Press the letter of the game you wish to enter.": {TokenUserPrompt, menuOnly},
	}

	for pattern, def := range patterns {
		l.patternMatchers[pattern] = &PatternMatcher{
			pattern:   pattern,
			tokenType: def.tokenType,
			states:    def.states,
		}
	}
}

// ProcessChunk processes raw data chunks for game detection
func (l *GameDetector) ProcessChunk(data []byte) {
	if len(data) == 0 {
		return
	}

	text := string(data)
	l.ProcessLine(text)
}

// ProcessUserInput processes user input separately from server output
// This should be called when the user types something, not for server output
func (l *GameDetector) ProcessUserInput(input string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Update activity timestamp
	l.lastActivity = time.Now()

	// Process isolated letters from user input in game menu state
	currentState := l.state.Load()
	trimmed := strings.TrimRight(input, "\r\n")
	if currentState.currentState == StateGameMenuVisible && len(trimmed) == 1 {
		char := rune(trimmed[0])
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') {
			letterStr := strings.ToUpper(string(char))
			// Use the game name from parsed options, or fall back to the letter itself
			gameName := letterStr
			if name, exists := currentState.gameOptions[letterStr]; exists {
				gameName = name
			}
			l.updateState(func(s *gameDetectorState) *gameDetectorState {
				newState := copyState(s)
				newState.selectedGame = gameName
				newState.selectedLetter = letterStr
				newState.currentState = StateGameSelected
				return newState
			})
		}
	}
}

// ProcessLine processes server output character by character with minimal buffering
func (l *GameDetector) ProcessLine(text string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Update activity timestamp
	l.lastActivity = time.Now()

	// Strip ANSI codes before processing using streaming stripper
	cleanText := l.ansiStripper.StripChunk(text)

	// Maintain recent content buffer for context analysis (keep last ~500 chars)
	l.recentContent += cleanText
	if len(l.recentContent) > 500 {
		l.recentContent = l.recentContent[len(l.recentContent)-500:]
	}

	// Process each character
	for _, char := range cleanText {
		l.processCharacter(char)
	}

	// Process any emitted tokens
	l.processTokens()

	// Check for timeout
	l.checkTimeout()
}

// processCharacter handles a single character through state-appropriate pattern matchers
func (l *GameDetector) processCharacter(char rune) {
	currentState := l.state.Load()

	// Check all pattern matchers that are active in the current state
	for _, matcher := range l.patternMatchers {
		if matcher.states[currentState.currentState] {
			l.updatePatternMatcher(matcher, char)
		}
	}

	// State-specific game option parsing (not pattern-based)
	switch currentState.currentState {
	case StateIdle, StateGameMenuVisible:
		if currentState.currentState == StateGameMenuVisible {
			// Check isolated letters before option matchers, since matchers
			// will consume the letter and become active
			l.processIsolatedLetter(char)
		}
		for _, m := range l.optionMatchers {
			if letter, name, ok := m.ProcessChar(char); ok {
				l.emitGameOptionToken(letter, name)
			}
		}
	}
}

// checkPattern checks a specific pattern matcher
func (l *GameDetector) checkPattern(pattern string, char rune) {
	if matcher, exists := l.patternMatchers[pattern]; exists {
		l.updatePatternMatcher(matcher, char)
	}
}

// updatePatternMatcher updates a single pattern matcher with the new character
func (l *GameDetector) updatePatternMatcher(matcher *PatternMatcher, char rune) {
	expectedChar := rune(matcher.pattern[matcher.position])

	if char == expectedChar {
		// Character matches, advance in pattern
		matcher.position++
		matcher.buffer += string(char)
		matcher.isActive = true

		// Check if we've matched the complete pattern
		if matcher.position >= len(matcher.pattern) {
			// Emit token
			token := Token{
				Type:  matcher.tokenType,
				Value: matcher.buffer,
				Pos:   0, // Position tracking simplified for streaming
			}
			l.tokens <- token

			// Reset matcher
			matcher.position = 0
			matcher.buffer = ""
			matcher.isActive = false
		}
	} else if matcher.isActive {
		// Pattern broken, reset but check if current char starts a new match
		matcher.position = 0
		matcher.buffer = ""
		matcher.isActive = false

		// Check if current character starts this pattern
		if char == rune(matcher.pattern[0]) {
			matcher.position = 1
			matcher.buffer = string(char)
			matcher.isActive = true
		}
	} else if char == rune(matcher.pattern[0]) {
		// Start of potential pattern match
		matcher.position = 1
		matcher.buffer = string(char)
		matcher.isActive = true
	}
}

// isolatedLetterState tracks context for isolated letter detection
type isolatedLetterState struct {
	prevChar     rune
	prevPrevChar rune
}

// processIsolatedLetter handles isolated letters from server output (could be echoed user input)
func (l *GameDetector) processIsolatedLetter(char rune) {
	// Store the current previous character before updating
	currentPrevChar := l.iLetterState.prevChar

	// Update state tracking
	l.iLetterState.prevPrevChar = l.iLetterState.prevChar
	l.iLetterState.prevChar = char

	// Only process if we're in game menu state and have game options
	currentState := l.state.Load()
	if currentState.currentState != StateGameMenuVisible {
		return
	}

	// Skip isolated letter detection if we're currently parsing a game option pattern
	// This prevents letters inside <A> patterns from being treated as user input
	for _, m := range l.optionMatchers {
		if m.IsActive() {
			return
		}
	}

	// Check if this is an isolated letter (A-Z)
	if char >= 'A' && char <= 'Z' {
		letterStr := string(char)

		// If we're expecting user input (just saw a prompt), accept any letter
		if currentState.expectingUserInput {
			l.emitIsolatedLetterToken(letterStr)
			return
		}

		// Otherwise require the letter to match a parsed game option
		if _, exists := currentState.gameOptions[letterStr]; exists {
			if l.isValidIsolatedLetterContext(currentPrevChar) {
				l.emitIsolatedLetterToken(letterStr)
			}
		}
	}
}

// isValidIsolatedLetterContext checks if the context is appropriate for isolated letter detection
func (l *GameDetector) isValidIsolatedLetterContext(prevChar rune) bool {
	// Strategy: Only accept isolated letters if we've recently detected a user prompt
	// This covers both direct responses (after colon) and echoed input (after various chars)

	// First, reject letters that are clearly part of game option patterns
	// Check if the previous character is '<' which indicates this is part of <X> pattern
	if prevChar == '<' {
		return false
	}

	// Check if we recently saw a user prompt - this is the key gate
	recentContext := l.recentContent
	if len(recentContext) > 50 {
		recentContext = recentContext[len(recentContext)-50:] // Check last 50 chars for prompts
	}

	// Look for user prompt patterns that indicate we're expecting user input
	promptIndicators := []string{
		"choice:", "selection:", "enter", "your choice", "please enter",
		"choice :", "selection :", // Handle spacing variations
		"select a game", // Game menu prompt
	}

	hasRecentPrompt := false
	for _, indicator := range promptIndicators {
		if strings.Contains(strings.ToLower(recentContext), indicator) {
			hasRecentPrompt = true
			break
		}
	}

	// If no recent prompt, reject the letter
	if !hasRecentPrompt {
		return false
	}

	// If we have a recent prompt, accept letters in reasonable contexts:
	// - After colons (direct response: "choice: A")
	// - After newlines (echoed input on new line: "\nA")
	// - At start of input
	// - After spaces only if very recent prompt (within last 10 chars)
	if prevChar == ':' || prevChar == '\n' || prevChar == '\r' || prevChar == 0 {
		return true
	}

	// For spaces, be more restrictive - only if prompt is very recent
	if prevChar == ' ' || prevChar == '\t' {
		// Check if we have a prompt in the last 10 characters (very recent)
		recentContext := l.recentContent
		if len(recentContext) > 10 {
			recentContext = recentContext[len(recentContext)-10:]
		}

		for _, indicator := range promptIndicators {
			if strings.Contains(strings.ToLower(recentContext), indicator) {
				return true
			}
		}
	}

	return false
}

// isValidPrecedingChar checks if the preceding character is valid for isolated letters (legacy)
func (l *GameDetector) isValidPrecedingChar(prevChar rune) bool {
	validChars := []rune{' ', ':', '>', '\n', '\r', '\t', '[', '(', ')'}
	for _, validChar := range validChars {
		if prevChar == validChar {
			return true
		}
	}
	return prevChar == 0 // Start of input
}

// emitGameOptionToken emits a game option token
func (l *GameDetector) emitGameOptionToken(letter, gameName string) {
	cleanGameName := strings.TrimSpace(gameName)

	// Auto-transition to game menu state if we detect game options while idle
	l.updateState(func(s *gameDetectorState) *gameDetectorState {
		if s.currentState == StateIdle {
			newState := copyState(s)
			newState.currentState = StateGameMenuVisible
			// Only initialize gameOptions if it's nil/empty to preserve existing options
			if newState.gameOptions == nil {
				newState.gameOptions = make(map[string]string)
			}
			return newState
		}
		return s
	})

	token := Token{
		Type:     TokenGameOption,
		Value:    "<" + letter + "> " + cleanGameName,
		Letter:   letter,
		GameName: cleanGameName,
		Pos:      0,
	}
	l.tokens <- token
}

// emitIsolatedLetterToken emits an isolated letter token
func (l *GameDetector) emitIsolatedLetterToken(letter string) {
	token := Token{
		Type:  TokenIsolatedLetter,
		Value: letter,
		Pos:   0,
	}
	l.tokens <- token
}

// processTokens handles emitted tokens and updates game detection state
func (l *GameDetector) processTokens() {
	for {
		select {
		case token := <-l.tokens:
			l.handleToken(token)
		default:
			return // No more tokens
		}
	}
}

// handleToken processes a single token and updates detection state
func (l *GameDetector) handleToken(token Token) {
	switch token.Type {
	case TokenGameMenu:
		// Detect game menu regardless of current state (could be returning from a game)
		l.updateState(func(s *gameDetectorState) *gameDetectorState {
			newState := copyState(s)
			newState.currentState = StateGameMenuVisible
			// Only reset gameOptions if we don't already have options (preserve auto-detected ones)
			if len(newState.gameOptions) == 0 {
				newState.gameOptions = make(map[string]string)
			}
			return newState
		})

	case TokenGameOption:
		l.updateState(func(s *gameDetectorState) *gameDetectorState {
			if s.currentState == StateGameMenuVisible {
				newState := copyState(s)
				newState.gameOptions[token.Letter] = token.GameName
				return newState
			}
			return s
		})

	case TokenIsolatedLetter:
		l.updateState(func(s *gameDetectorState) *gameDetectorState {
			if s.currentState == StateGameMenuVisible {
				newState := copyState(s)
				newState.selectedLetter = token.Value
				newState.expectingUserInput = false
				if gameName, exists := s.gameOptions[token.Value]; exists {
					newState.selectedGame = gameName
				} else {
					newState.selectedGame = token.Value
				}
				newState.currentState = StateGameSelected
				return newState
			}
			return s
		})

	case TokenGameStart:
		l.updateState(func(s *gameDetectorState) *gameDetectorState {
			if s.currentState == StateGameSelected || s.currentState == StateGameMenuVisible {
				newState := copyState(s)
				newState.currentState = StateGameActive
				return newState
			}
			return s
		})
		// Load database only if we know which game was selected
		currentState := l.state.Load()
		if currentState.currentState == StateGameActive && currentState.selectedLetter != "" {
			if err := l.loadGameDatabase(); err != nil {
				log.Info("GameDetector: loadGameDatabase() failed", "error", err)
			}
		}

	case TokenGameExit:
		currentState := l.state.Load()
		if currentState.currentState == StateGameActive || currentState.currentState == StateGameSelected {
			l.resetGameState()
		}

	case TokenMainMenu:
		// "TWGS v" or "TradeWars Game Server" detected
		// Ignore in GameActive — can appear in game content (config screens, etc.)
		currentState := l.state.Load()
		if currentState.currentState != StateGameActive {
			l.updateState(func(s *gameDetectorState) *gameDetectorState {
				newState := copyState(s)
				newState.currentState = StateGameMenuVisible
				if newState.gameOptions == nil {
					newState.gameOptions = make(map[string]string)
				}
				return newState
			})
		}

	case TokenUserPrompt:
		// A user prompt was detected - we're now expecting user input
		l.updateState(func(s *gameDetectorState) *gameDetectorState {
			newState := copyState(s)
			newState.expectingUserInput = true
			return newState
		})
	}
}

// resetGameState clears current game state
// This method assumes the caller already holds the mutex lock for non-atomic operations
func (l *GameDetector) resetGameState() {
	// Get current state for database cleanup
	oldState := l.state.Load()

	// Notify about database being unloaded if one is currently loaded
	if l.currentDatabase != nil && l.onDatabaseStateChanged != nil {
		currentGame := oldState.selectedGame
		if currentGame == "" {
			currentGame = "Unknown Game"
		}
		currentDbName := l.createDatabaseName()
		go func() {
			l.onDatabaseStateChanged(currentGame, l.serverHost, l.serverPort, currentDbName, false)
		}()
	}

	// Atomically reset the state
	newState := &gameDetectorState{
		currentState:       StateIdle,
		selectedGame:       "",
		gameOptions:        make(map[string]string),
		expectingUserInput: false,
	}
	l.state.Store(newState)

	// Reset non-atomic state (assumes caller holds mutex)
	l.recentContent = "" // Clear recent content buffer

	// Reset pattern matchers
	for _, matcher := range l.patternMatchers {
		matcher.position = 0
		matcher.buffer = ""
		matcher.isActive = false
	}

	// Reset state machines
	for _, m := range l.optionMatchers {
		m.Reset()
	}
	l.iLetterState.prevChar = 0
	l.iLetterState.prevPrevChar = 0

	// Reset ANSI stripper state
	l.ansiStripper.Reset()
}

// Helper functions

// isLikelyGameContent determines if a main menu pattern is likely appearing
// within game content rather than as an actual menu transition
func (l *GameDetector) isLikelyGameContent(pattern string) bool {
	// Use the recent content buffer for context analysis
	recentContext := l.recentContent

	// Look for indicators that we're in game content:
	// - Version info patterns ("Ver#", "running under")
	// - Game stats patterns ("Stats for", "ports are open", "Traders")
	// - Command prompts ("Command [", "?=Help")
	// - Configuration screens ("Configuration", "Status")

	gameContentIndicators := []string{
		"Ver#", "running under", "registered to",
		"Stats for", "ports are open", "Traders", "business",
		"Command [", "?=Help", "(?=Help)",
		"Configuration", "Status", "days",
		"Universe", "Corporations", "Fighters",
		"sectors", "planets", "Citadels", "net worth",
	}

	// Check if any recent content suggests we're viewing game information
	for _, indicator := range gameContentIndicators {
		if strings.Contains(recentContext, indicator) {
			return true
		}
	}

	// Additional heuristic: if the pattern is "TWGS v" in certain contexts,
	// it's likely version info rather than a menu header
	if strings.HasPrefix(pattern, "TWGS v") {
		// If it's "TWGS v" followed by a number > 2, or contains "running under", it's version info
		if strings.Contains(recentContext, "running under") ||
			strings.Contains(recentContext, "Ver#") {
			return true
		}
	}

	return false
}

// isAlphaNum checks if character is alphanumeric
func (l *GameDetector) isAlphaNum(char byte) bool {
	return (char >= 'A' && char <= 'Z') ||
		(char >= 'a' && char <= 'z') ||
		(char >= '0' && char <= '9')
}

// checkTimeout resets state if no activity for too long
func (l *GameDetector) checkTimeout() {
	if time.Since(l.lastActivity) > l.detectionTimeout {
		l.resetGameState()
	}
}

// CheckTimeoutManual allows manual timeout checking for tests
func (l *GameDetector) CheckTimeoutManual() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.checkTimeout()
}

// resetGameState clears current game state

// Database management functions (reused from original)

func (l *GameDetector) loadGameDatabase() error {
	// Notify about database being unloaded if one is currently loaded
	if l.currentDatabase != nil && l.onDatabaseStateChanged != nil {
		currentState := l.state.Load()
		currentGame := currentState.selectedGame
		if currentGame == "" {
			// If we're replacing a database, use the previous game name if available
			currentGame = "Unknown Game"
		}
		currentDbName := l.createDatabaseName()
		go func() {
			l.onDatabaseStateChanged(currentGame, l.serverHost, l.serverPort, currentDbName, false)
		}()
	}

	if l.currentDatabase != nil {
		l.currentDatabase.CloseDatabase()
	}

	if l.currentScriptManager != nil {
		l.currentScriptManager.Stop()
	}

	currentState := l.state.Load()
	dbName := l.createDatabaseName()

	log.Info("GAME DETECTOR: Loading database", "dbName", dbName, "selectedGame", currentState.selectedGame)

	db := database.NewDatabase()

	if err := db.CreateDatabase(dbName); err != nil {
		if err := db.OpenDatabase(dbName); err != nil {
			return fmt.Errorf("failed to load database %s: %w", dbName, err)
		}
	}

	log.Info("GAME DETECTOR: Successfully loaded database", "dbName", dbName)

	scriptManager := scripting.NewScriptManager(db)

	l.currentDatabase = db
	l.currentScriptManager = scriptManager

	// Notify about database state change (loaded)
	if l.onDatabaseStateChanged != nil {
		go func() {
			l.onDatabaseStateChanged(currentState.selectedGame, l.serverHost, l.serverPort, dbName, true)
		}()
	} else {
	}

	if l.onDatabaseLoaded != nil {
		log.Info("GameDetector: triggering onDatabaseLoaded callback", "db", db)
		go func() {
			if err := l.onDatabaseLoaded(db, scriptManager); err != nil {
				log.Info("GameDetector: onDatabaseLoaded callback error", "error", err)
			} else {
				log.Info("GameDetector: onDatabaseLoaded callback completed successfully")
			}
		}()
	} else {
		log.Info("GameDetector: no onDatabaseLoaded callback set")
	}

	return nil
}

func (l *GameDetector) createDatabaseName() string {
	currentState := l.state.Load()
	host := sanitizeForFilename(l.serverHost)
	port := sanitizeForFilename(l.serverPort)
	letter := strings.ToLower(currentState.selectedLetter)
	if letter == "" {
		letter = "default"
	}
	return fmt.Sprintf("%s_%s_%s.db", host, port, letter)
}

// sanitizeForFilename removes or replaces characters that are not safe for filenames
func sanitizeForFilename(input string) string {
	// Replace unsafe characters with underscores
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " ", "."}
	result := input
	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Remove consecutive underscores
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	// Trim underscores from start and end
	result = strings.Trim(result, "_")
	return strings.ToLower(result)
}

// reset is an alias for resetGameState for test compatibility
func (l *GameDetector) reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.resetGameState()
}

// Public interface methods (matching original GameDetector)

func (l *GameDetector) SetDetectionTimeout(timeout time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.detectionTimeout = timeout
}

func (l *GameDetector) SetDatabaseLoadedCallback(callback func(db database.Database, scriptManager *scripting.ScriptManager) error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.onDatabaseLoaded = callback
}

func (l *GameDetector) SetDatabaseStateChangedCallback(callback func(gameName, serverHost, serverPort, dbName string, isLoaded bool)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.onDatabaseStateChanged = callback
}

func (l *GameDetector) GetCurrentGame() string {
	state := l.state.Load()
	if state == nil {
		return ""
	}
	return state.selectedGame
}

func (l *GameDetector) GetState() GameDetectionState {
	state := l.state.Load()
	if state == nil {
		return StateIdle
	}
	return state.currentState
}

func (l *GameDetector) IsGameActive() bool {
	state := l.state.Load()
	if state == nil {
		return false
	}
	return state.currentState == StateGameActive
}

func (l *GameDetector) GetCurrentDatabase() database.Database {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.currentDatabase
}

func (l *GameDetector) GetCurrentScriptManager() *scripting.ScriptManager {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.currentScriptManager
}

func (l *GameDetector) GetServerInfo() ConnectionInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return ConnectionInfo{
		Host: l.serverHost,
		Port: l.serverPort,
	}
}

func (l *GameDetector) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Notify about database being unloaded
	if l.currentDatabase != nil && l.onDatabaseStateChanged != nil {
		currentState := l.state.Load()
		currentGame := currentState.selectedGame
		if currentGame == "" {
			currentGame = "Unknown Game"
		}
		currentDbName := l.createDatabaseName()
		go func() {
			l.onDatabaseStateChanged(currentGame, l.serverHost, l.serverPort, currentDbName, false)
		}()
	}

	if l.currentScriptManager != nil {
		l.currentScriptManager.Stop()
	}

	if l.currentDatabase != nil {
		return l.currentDatabase.CloseDatabase()
	}

	return nil
}
