package api

import (
	"errors"
	coreapi "twist/internal/api"
	"twist/internal/api/factory"
)

// ProxyClient manages ProxyAPI connections for TUI
type ProxyClient struct {
	currentAPI coreapi.ProxyAPI // Current active connection (nil if disconnected)
}

// NewProxyClient creates a new proxy client
func NewProxyClient() *ProxyClient {
	return &ProxyClient{
		currentAPI: nil,
	}
}

func (pc *ProxyClient) Connect(address string, tuiAPI coreapi.TuiAPI, opts ...*coreapi.ConnectOptions) error {
	var connectOpts *coreapi.ConnectOptions
	if len(opts) > 0 {
		connectOpts = opts[0]
	}
	proxyAPI := factory.Connect(address, tuiAPI, connectOpts)

	// Store the connected API instance
	pc.currentAPI = proxyAPI
	return nil
}

func (pc *ProxyClient) Disconnect() error {
	if pc.currentAPI == nil {
		return nil
	}

	err := pc.currentAPI.Disconnect()
	pc.currentAPI = nil // Clear reference after disconnect
	return err
}

func (pc *ProxyClient) IsConnected() bool {
	if pc.currentAPI == nil {
		return false
	}
	return pc.currentAPI.IsConnected()
}

func (pc *ProxyClient) SendData(data []byte) error {
	if pc.currentAPI == nil {
		return errors.New("not connected")
	}
	return pc.currentAPI.SendData(data)
}

func (pc *ProxyClient) GetCurrentAPI() coreapi.ProxyAPI {
	return pc.currentAPI
}
