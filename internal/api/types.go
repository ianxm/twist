package api

// ConnectOptions holds optional parameters for Connect
type ConnectOptions struct {
	DatabasePath  string
	ScriptManager interface{} // Existing *scripting.ScriptManager to adopt (interface{} to avoid import cycle)
}
