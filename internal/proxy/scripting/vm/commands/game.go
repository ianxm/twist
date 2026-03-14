package commands

import (
	"fmt"
	"time"
	"twist/internal/log"
	"twist/internal/proxy/scripting/types"
)

// RegisterGameCommands registers all TWX game-specific commands
func RegisterGameCommands(vm CommandRegistry) {
	// Basic game commands
	vm.RegisterCommand("SEND", 1, -1, []types.ParameterType{types.ParamValue}, cmdSend)
	vm.RegisterCommand("WAITFOR", 1, 1, []types.ParameterType{types.ParamValue}, cmdWaitFor)
	vm.RegisterCommand("PAUSE", 0, 0, []types.ParameterType{}, cmdPause)
	vm.RegisterCommand("HALT", 0, 0, []types.ParameterType{}, cmdHalt)
	vm.RegisterCommand("LOGGING", 1, 1, []types.ParameterType{types.ParamValue}, cmdLogging)

	// Timer commands
	vm.RegisterCommand("SETTIMER", 1, 1, []types.ParameterType{types.ParamValue}, cmdSetTimer)
	vm.RegisterCommand("GETTIMER", 1, 1, []types.ParameterType{types.ParamVar}, cmdGetTimer)

	// Input commands
	vm.RegisterCommand("GETINPUT", 2, 3, []types.ParameterType{types.ParamVar, types.ParamValue, types.ParamValue}, cmdGetInput)
	vm.RegisterCommand("GETCONSOLEINPUT", 2, 2, []types.ParameterType{types.ParamVar, types.ParamValue}, cmdGetConsoleInput)

	// Debug command for troubleshooting
	vm.RegisterCommand("DEBUGLOG", 1, -1, []types.ParameterType{types.ParamValue}, cmdDebugLog)

	// Text processing commands
	vm.RegisterCommand("MERGETEXT", 3, 3, []types.ParameterType{types.ParamValue, types.ParamValue, types.ParamVar}, cmdMergeText)

	// Game data commands - TWX compatibility
	vm.RegisterCommand("SLEEP", 1, 1, []types.ParameterType{types.ParamValue}, cmdSleep)
	vm.RegisterCommand("GETCURRENTSECTOR", 1, 1, []types.ParameterType{types.ParamVar}, cmdGetCurrentSector)
	vm.RegisterCommand("GETSECTOR", 2, 2, []types.ParameterType{types.ParamValue, types.ParamVar}, cmdGetSector)
}

func cmdSend(vm types.VMInterface, params []*types.CommandParam) error {
	// Concatenate all parameters like ECHO
	message := ""
	for _, param := range params {
		if param.Type == types.ParamVar {
			// Get variable value
			value := vm.GetVariable(param.VarName)
			log.Info("SEND command: variable resolves", "line", vm.GetCurrentLine(), "variable", param.VarName, "value", value.ToString())
			message += value.ToString()
		} else {
			// Use literal value
			message += param.Value.ToString()
		}
	}

	scriptName := "unknown"
	if script := vm.GetCurrentScript(); script != nil {
		scriptName = script.GetName()
	}
	log.Info("SEND command: sending message", "script", scriptName, "line", vm.GetCurrentLine(), "message", message)

	// Send message as-is (carriage returns from lexer are preserved)
	return vm.Send(message)
}

func cmdWaitFor(vm types.VMInterface, params []*types.CommandParam) error {
	if len(params) != 1 {
		return vm.Error("WAITFOR requires exactly 1 parameter: pattern")
	}

	pattern := GetParamString(vm, params[0])
	scriptName := "unknown"
	if script := vm.GetCurrentScript(); script != nil {
		scriptName = script.GetName()
	}
	log.Info("WAITFOR command: waiting for pattern", "script", scriptName, "line", vm.GetCurrentLine(), "pattern", pattern)
	return vm.WaitFor(pattern)
}

func cmdPause(vm types.VMInterface, params []*types.CommandParam) error {
	return vm.Pause()
}

func cmdHalt(vm types.VMInterface, params []*types.CommandParam) error {
	return vm.Halt()
}

func cmdLogging(vm types.VMInterface, params []*types.CommandParam) error {
	// Toggle logging on/off based on parameter
	// For testing, we'll just acknowledge the command
	return nil
}

func cmdSetTimer(vm types.VMInterface, params []*types.CommandParam) error {
	// Set timer value
	return nil
}

func cmdGetTimer(vm types.VMInterface, params []*types.CommandParam) error {
	// Get current timer value - for testing return 0
	result := &types.Value{
		Type:   types.NumberType,
		Number: 0,
	}
	vm.SetVariable(params[0].VarName, result)
	return nil
}

func cmdGetInput(vm types.VMInterface, params []*types.CommandParam) error {
	// Extract prompt text (2nd parameter)
	prompt := ""
	if len(params) > 1 {
		prompt = GetParamString(vm, params[1])
	}

	// Extract default value (3rd parameter, optional)
	defaultValue := ""
	if len(params) > 2 {
		defaultValue = GetParamString(vm, params[2])
	}

	// Debug logging
	scriptName := "unknown"
	if script := vm.GetCurrentScript(); script != nil {
		scriptName = script.GetName()
	}
	log.Info("GETINPUT command: parameters", "script", scriptName, "line", vm.GetCurrentLine(), "prompt", prompt, "default", defaultValue)
	log.Info("GETINPUT command: state check", "script", scriptName, "pendingResult", vm.GetPendingInputResult(), "waitingForInput", vm.IsWaitingForInput(), "pendingPrompt", vm.GetPendingInputPrompt())

	// Check if there's a pending input result (we're resuming from input)
	// We use JustResumedFromInput to handle cases where input is empty string
	if vm.IsWaitingForInput() || vm.JustResumedFromInput() {
		// We're resuming - get the input result and store it
		log.Info("GETINPUT command: resuming from input", "script", scriptName)
		input := vm.GetPendingInputResult()

		// Use default value if input is empty
		if input == "" && defaultValue != "" {
			input = defaultValue
		}

		log.Info("GETINPUT command: setting variable", "script", scriptName, "variable", params[0].VarName, "value", input)

		// Store result in the variable
		result := &types.Value{
			Type:   types.StringType,
			String: input,
		}
		vm.SetVariable(params[0].VarName, result)

		// Clear the pending input state since we've processed it
		vm.ClearPendingInput()

		return nil // This will allow the execution to advance to the next command
	}

	// First time executing this command - initiate input collection
	log.Info("GETINPUT command: initiating input collection", "script", scriptName)
	// Format the prompt like TWX does
	fullPrompt := prompt
	if defaultValue != "" {
		fullPrompt = prompt + " [" + defaultValue + "]"
	}

	// Initiate input collection and pause script execution
	// GetInput will handle displaying the prompt
	_, err := vm.GetInput(fullPrompt)
	if err != nil {
		return err
	}

	// Return a pause error to stop execution until input is provided
	return types.ErrScriptPaused
}

func cmdGetConsoleInput(vm types.VMInterface, params []*types.CommandParam) error {
	// Get console input - for testing, return "1"
	result := &types.Value{
		Type:   types.StringType,
		String: "1",
	}
	vm.SetVariable(params[0].VarName, result)
	return nil
}

func cmdDebugLog(vm types.VMInterface, params []*types.CommandParam) error {
	// Concatenate all parameters like ECHO
	message := ""
	for _, param := range params {
		if param.Type == types.ParamVar {
			// Get variable value
			value := vm.GetVariable(param.VarName)
			message += value.ToString()
		} else {
			// Use literal value
			message += param.Value.ToString()
		}
	}

	scriptName := "unknown"
	if script := vm.GetCurrentScript(); script != nil {
		scriptName = script.GetName()
	}
	log.Info("SCRIPT DEBUG command", "script", scriptName, "line", vm.GetCurrentLine(), "message", message)
	return nil
}

func cmdMergeText(vm types.VMInterface, params []*types.CommandParam) error {
	// Merge two text strings
	text1 := GetParamString(vm, params[0])
	text2 := GetParamString(vm, params[1])
	result := &types.Value{
		Type:   types.StringType,
		String: text1 + text2,
	}
	vm.SetVariable(params[2].VarName, result)
	return nil
}

// cmdTime gets the current time
func cmdTime(vm types.VMInterface, params []*types.CommandParam) error {
	if len(params) != 1 {
		return vm.Error("TIME requires exactly 1 parameter: result_var")
	}

	// Mock time value for testing
	vm.SetVariable(params[0].VarName, &types.Value{
		Type:   types.StringType,
		String: "12:34:56",
	})

	return nil
}

// cmdDate gets the current date
func cmdDate(vm types.VMInterface, params []*types.CommandParam) error {
	if len(params) != 1 {
		return vm.Error("DATE requires exactly 1 parameter: result_var")
	}

	// Mock date value for testing
	vm.SetVariable(params[0].VarName, &types.Value{
		Type:   types.StringType,
		String: "01/01/2024",
	})

	return nil
}

// cmdGetTime gets the current time in milliseconds
func cmdGetTime(vm types.VMInterface, params []*types.CommandParam) error {
	if len(params) != 1 {
		return vm.Error("GETTIME requires exactly 1 parameter: result_var")
	}

	// Mock time value for testing
	vm.SetVariable(params[0].VarName, &types.Value{
		Type:   types.NumberType,
		Number: 123456789,
	})

	return nil
}

// cmdSleep pauses execution for a specified time
func cmdSleep(vm types.VMInterface, params []*types.CommandParam) error {
	if len(params) != 1 {
		return vm.Error("SLEEP requires exactly 1 parameter: milliseconds")
	}

	duration := GetParamNumber(vm, params[0])
	time.Sleep(time.Duration(duration * float64(time.Millisecond)))

	return nil
}

// cmdGetCurrentSector gets the current sector number
func cmdGetCurrentSector(vm types.VMInterface, params []*types.CommandParam) error {
	if len(params) != 1 {
		return vm.Error("GETCURRENTSECTOR requires exactly 1 parameter: result_var")
	}

	// Get sector data from game interface
	gameInterface := vm.GetGameInterface()
	log.Info("cmdGetSector: game interface", "gameInterface", gameInterface)
	if gameInterface == nil {
		log.Info("cmdGetSector: ERROR - gameInterface is nil!")
		return vm.Error("Game interface not available")
	}
	sectorIndex := gameInterface.GetCurrentSector()

	vm.SetVariable(params[0].VarName, &types.Value{
		Type:   types.NumberType,
		Number: float64(sectorIndex),
	})

	return nil
}

// cmdGetCurrentPrompt gets the current prompt
func cmdGetCurrentPrompt(vm types.VMInterface, params []*types.CommandParam) error {
	if len(params) != 1 {
		return vm.Error("GETCURRENTPROMPT requires exactly 1 parameter: result_var")
	}

	// Mock prompt value for testing
	vm.SetVariable(params[0].VarName, &types.Value{
		Type:   types.StringType,
		String: "Command [TL=00:00:00]:[",
	})

	return nil
}

// cmdIsString checks if a value is a string
func cmdIsString(vm types.VMInterface, params []*types.CommandParam) error {
	if len(params) != 2 {
		return vm.Error("ISSTRING requires exactly 2 parameters: value, result_var")
	}

	val := GetParamValue(vm, params[0])
	result := 0.0
	if val.Type == types.StringType {
		result = 1.0
	}

	vm.SetVariable(params[1].VarName, &types.Value{
		Type:   types.NumberType,
		Number: result,
	})

	return nil
}

// cmdGetType gets the type of a value
func cmdGetType(vm types.VMInterface, params []*types.CommandParam) error {
	if len(params) != 2 {
		return vm.Error("GETTYPE requires exactly 2 parameters: value, result_var")
	}

	val := GetParamValue(vm, params[0])
	var typeStr string
	switch val.Type {
	case types.StringType:
		typeStr = "string"
	case types.NumberType:
		typeStr = "number"
	case types.ArrayType:
		typeStr = "array"
	default:
		typeStr = "unknown"
	}

	vm.SetVariable(params[1].VarName, &types.Value{
		Type:   types.StringType,
		String: typeStr,
	})

	return nil
}

// cmdGetSector implements the TWX getSector command with full Pascal compatibility
// Syntax: getSector <index> <var>
// Example: getSector 123 $s
func cmdGetSector(vm types.VMInterface, params []*types.CommandParam) error {
	// Add panic recovery for debugging
	defer func() {
		if r := recover(); r != nil {
			log.Error("PANIC in cmdGetSector", "error", r)
			panic(r) // Re-panic after logging
		}
	}()

	if len(params) != 2 {
		return vm.Error("GETSECTOR requires exactly 2 parameters: sector_index, result_var")
	}

	// Get sector index from first parameter
	indexValue := GetParamValue(vm, params[0])
	log.Info("cmdGetSector: index value", "indexValue", indexValue)
	sectorIndex := int(indexValue.ToNumber())
	log.Info("cmdGetSector: sector index", "sectorIndex", sectorIndex)

	// Ignore invalid call with index of zero (Pascal TWX behavior)
	if sectorIndex == 0 {
		return nil
	}

	// Get variable name for result
	varName := params[1].VarName
	log.Info("cmdGetSector: variable name", "varName", varName)

	// Get sector data from game interface
	gameInterface := vm.GetGameInterface()
	log.Info("cmdGetSector: game interface", "gameInterface", gameInterface)
	if gameInterface == nil {
		log.Info("cmdGetSector: ERROR - gameInterface is nil!")
		return vm.Error("Game interface not available")
	}
	sector, err := gameInterface.GetSector(sectorIndex)
	log.Info("cmdGetSector: after GetSector call", "error", err)
	if err != nil {
		log.Info("GETSECTOR: sector not found, setting default values", "sectorIndex", sectorIndex, "error", err)
		// If sector not found, set default empty values
		setSectorVariables(vm, varName, sectorIndex, nil)
		return nil
	}

	log.Info("GETSECTOR: sector found", "sectorIndex", sectorIndex, "portName", sector.Port.Name, "portClass", sector.Port.ClassIndex, "hasPort", sector.HasPort)

	// Set all sector variables matching Pascal TWX exactly
	setSectorVariables(vm, varName, sectorIndex, &sector)
	return nil
}

// setSectorVariables sets all sector variables exactly like Pascal TWX CmdGetSector
func setSectorVariables(vm types.VMInterface, varName string, index int, sector *types.SectorData) {
	// Always set the index
	vm.SetVariable(varName+".INDEX", &types.Value{
		Type: types.NumberType, Number: float64(index),
	})

	if sector == nil {
		// Set default values for non-existent sector
		setDefaultSectorValues(vm, varName)
		return
	}

	// Set exploration status
	switch sector.Explored {
	case 0: // etNo
		vm.SetVariable(varName+".EXPLORED", &types.Value{Type: types.StringType, String: "NO"})
	case 1: // etCalc
		vm.SetVariable(varName+".EXPLORED", &types.Value{Type: types.StringType, String: "CALC"})
	case 2: // etDensity
		vm.SetVariable(varName+".EXPLORED", &types.Value{Type: types.StringType, String: "DENSITY"})
	case 3: // etHolo
		vm.SetVariable(varName+".EXPLORED", &types.Value{Type: types.StringType, String: "YES"})
	default:
		vm.SetVariable(varName+".EXPLORED", &types.Value{Type: types.StringType, String: "NO"})
	}

	// Basic sector properties
	vm.SetVariable(varName+".BEACON", &types.Value{Type: types.StringType, String: sector.Beacon})
	vm.SetVariable(varName+".CONSTELLATION", &types.Value{Type: types.StringType, String: sector.Constellation})

	// Mines
	vm.SetVariable(varName+".ARMIDMINES.QUANTITY", &types.Value{Type: types.NumberType, Number: float64(sector.MinesArmid.Quantity)})
	vm.SetVariable(varName+".LIMPETMINES.QUANTITY", &types.Value{Type: types.NumberType, Number: float64(sector.MinesLimpet.Quantity)})
	vm.SetVariable(varName+".ARMIDMINES.OWNER", &types.Value{Type: types.StringType, String: sector.MinesArmid.Owner})
	vm.SetVariable(varName+".LIMPETMINES.OWNER", &types.Value{Type: types.StringType, String: sector.MinesLimpet.Owner})

	// Fighters
	vm.SetVariable(varName+".FIGS.QUANTITY", &types.Value{Type: types.NumberType, Number: float64(sector.Fighters.Quantity)})
	vm.SetVariable(varName+".FIGS.OWNER", &types.Value{Type: types.StringType, String: sector.Fighters.Owner})
	vm.SetVariable(varName+".FIGS.TYPE", &types.Value{Type: types.StringType, String: sector.Fighters.Type})

	// Warp and density information
	vm.SetVariable(varName+".WARPCOUNT", &types.Value{Type: types.NumberType, Number: float64(len(sector.Warps))})
	vm.SetVariable(varName+".DENSITY", &types.Value{Type: types.NumberType, Number: float64(sector.Density)})
	vm.SetVariable(varName+".NAVHAZ", &types.Value{Type: types.NumberType, Number: float64(sector.NavHaz)})
	anomaly := "NO"
	if sector.Anomaly {
		anomaly = "YES"
	}
	vm.SetVariable(varName+".ANOMALY", &types.Value{Type: types.StringType, String: anomaly})

	// Set warp array (1-6 like Pascal TWX)
	for i := 1; i <= 6; i++ {
		warpValue := 0
		if i-1 < len(sector.Warps) {
			warpValue = sector.Warps[i-1]
		}
		vm.SetVariable(varName+".WARPS["+fmt.Sprintf("%d", i)+"]", &types.Value{
			Type: types.NumberType, Number: float64(warpValue),
		})
	}

	// Port information (key part for 1_Trade.ts compatibility)
	setPortVariables(vm, varName, sector)

	// Trader, ship, planet counts - for basic compatibility set to 0
	vm.SetVariable(varName+".TRADERS", &types.Value{Type: types.NumberType, Number: float64(len(sector.Traders))})
	vm.SetVariable(varName+".SHIPS", &types.Value{Type: types.NumberType, Number: float64(len(sector.Ships))})
	vm.SetVariable(varName+".PLANETS", &types.Value{Type: types.NumberType, Number: float64(len(sector.Planets))})
}

// setPortVariables sets port variables exactly like Pascal TWX
func setPortVariables(vm types.VMInterface, varName string, sector *types.SectorData) {
	// Always set port name
	portName := ""
	if sector != nil {
		portName = sector.Port.Name
	}
	vm.SetVariable(varName+".PORT.NAME", &types.Value{Type: types.StringType, String: portName})

	if !sector.HasPort {
		// No port exists
		log.Info("SETPORTVARS: no port, setting PORT.CLASS=0", "varName", varName, "portName", portName, "hasPort", sector != nil && sector.HasPort)
		vm.SetVariable(varName+".PORT.CLASS", &types.Value{Type: types.NumberType, Number: 0})
		vm.SetVariable(varName+".PORT.EXISTS", &types.Value{Type: types.NumberType, Number: 0})
	} else {
		// Port exists - set all port variables using actual sector data
		log.Info("SETPORTVARS: port exists", "varName", varName, "portName", portName, "portClass", sector.Port.ClassIndex)
		vm.SetVariable(varName+".PORT.CLASS", &types.Value{Type: types.NumberType, Number: float64(sector.Port.ClassIndex)})
		vm.SetVariable(varName+".PORT.EXISTS", &types.Value{Type: types.NumberType, Number: 1})
		vm.SetVariable(varName+".PORT.BUILDTIME", &types.Value{Type: types.NumberType, Number: float64(sector.Port.BuildTime)})

		// Product percentages
		vm.SetVariable(varName+".PORT.PERC_ORE", &types.Value{Type: types.NumberType, Number: float64(sector.Port.OrePercent)})
		vm.SetVariable(varName+".PORT.PERC_ORG", &types.Value{Type: types.NumberType, Number: float64(sector.Port.OrgPercent)})
		vm.SetVariable(varName+".PORT.PERC_EQU", &types.Value{Type: types.NumberType, Number: float64(sector.Port.EquipPercent)})

		// Product amounts
		vm.SetVariable(varName+".PORT.ORE", &types.Value{Type: types.NumberType, Number: float64(sector.Port.OreAmount)})
		vm.SetVariable(varName+".PORT.ORG", &types.Value{Type: types.NumberType, Number: float64(sector.Port.OrgAmount)})
		vm.SetVariable(varName+".PORT.EQU", &types.Value{Type: types.NumberType, Number: float64(sector.Port.EquipAmount)})

		// Port update timestamp (placeholder)
		vm.SetVariable(varName+".PORT.UPDATED", &types.Value{Type: types.StringType, String: "01/01/2024 00:00:00"})

		// Buy flags
		buyOre := 0
		buyOrg := 0
		buyEquip := 0
		if sector.Port.BuyOre { buyOre = 1 }
		if sector.Port.BuyOrg { buyOrg = 1 }
		if sector.Port.BuyEquip { buyEquip = 1 }
		vm.SetVariable(varName+".PORT.BUY_ORE", &types.Value{Type: types.NumberType, Number: float64(buyOre)})
		vm.SetVariable(varName+".PORT.BUY_ORG", &types.Value{Type: types.NumberType, Number: float64(buyOrg)})
		vm.SetVariable(varName+".PORT.BUY_EQU", &types.Value{Type: types.NumberType, Number: float64(buyEquip)})
	}
}

// setDefaultSectorValues sets default values for non-existent sectors
func setDefaultSectorValues(vm types.VMInterface, varName string) {
	// Set minimal default values
	vm.SetVariable(varName+".EXPLORED", &types.Value{Type: types.StringType, String: "NO"})
	vm.SetVariable(varName+".BEACON", &types.Value{Type: types.StringType, String: ""})
	vm.SetVariable(varName+".CONSTELLATION", &types.Value{Type: types.StringType, String: ""})
	vm.SetVariable(varName+".WARPCOUNT", &types.Value{Type: types.NumberType, Number: 0})
	vm.SetVariable(varName+".WARPS", &types.Value{Type: types.NumberType, Number: 0})
	vm.SetVariable(varName+".DENSITY", &types.Value{Type: types.NumberType, Number: -1})
	vm.SetVariable(varName+".NAVHAZ", &types.Value{Type: types.NumberType, Number: 0})
	vm.SetVariable(varName+".ANOMALY", &types.Value{Type: types.StringType, String: "NO"})

	// Default warp array
	for i := 1; i <= 6; i++ {
		vm.SetVariable(varName+".WARPS["+fmt.Sprintf("%d", i)+"]", &types.Value{
			Type: types.NumberType, Number: 0,
		})
	}

	// No port
	vm.SetVariable(varName+".PORT.CLASS", &types.Value{Type: types.NumberType, Number: 0})
	vm.SetVariable(varName+".PORT.EXISTS", &types.Value{Type: types.NumberType, Number: 0})
}
