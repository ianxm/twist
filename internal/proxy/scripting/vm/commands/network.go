package commands

import (
	"fmt"
	"twist/internal/proxy/scripting/types"
)

// RegisterNetworkCommands registers network commands with the VM
func RegisterNetworkCommands(vm CommandRegistry) {
	vm.RegisterCommand("CONNECT", 1, 2, []types.ParameterType{types.ParamValue, types.ParamValue}, cmdConnect)
	vm.RegisterCommand("DISCONNECT", 0, 0, []types.ParameterType{}, cmdDisconnect)
}

// cmdConnect initiates a proxy-level connection to a game server
// Accepts either: CONNECT "host:port" or CONNECT host port
func cmdConnect(vm types.VMInterface, params []*types.CommandParam) error {
	var address string
	if len(params) == 1 {
		address = GetParamString(vm, params[0])
	} else if len(params) == 2 {
		host := GetParamString(vm, params[0])
		port := GetParamString(vm, params[1])
		if host == "" || port == "" {
			return vm.Error("CONNECT requires non-empty host and port")
		}
		address = fmt.Sprintf("%s:%s", host, port)
	} else {
		return vm.Error("CONNECT requires 1 or 2 parameters: \"host:port\" or host port")
	}

	if address == "" {
		return vm.Error("CONNECT requires a non-empty address")
	}

	gameInterface := vm.GetGameInterface()
	if gameInterface == nil {
		return vm.Error("CONNECT: game interface not available")
	}

	if err := gameInterface.Connect(address); err != nil {
		return vm.Error(fmt.Sprintf("CONNECT failed: %v", err))
	}

	return nil
}

// cmdDisconnect disconnects from the game server
func cmdDisconnect(vm types.VMInterface, params []*types.CommandParam) error {
	gameInterface := vm.GetGameInterface()
	if gameInterface == nil {
		return vm.Error("DISCONNECT: game interface not available")
	}

	if err := gameInterface.Disconnect(); err != nil {
		return vm.Error(fmt.Sprintf("DISCONNECT failed: %v", err))
	}

	return nil
}
