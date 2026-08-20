package bridge

// Bridge IPC routes and frontend bindings.
// These functions wrap backend logic and are exposed to the frontend via Wails Bind.

// AppInfo returns application information.
func AppInfo() map[string]string {
	return map[string]string{
		"name":    "Conductino Study Browser",
		"version": "0.3.0",
		"engine":  "wails-v2",
	}
}

// Greet returns a greeting string.
func Greet(name string) string {
	if name == "" {
		name = "researcher"
	}
	return "Conductino is alive, " + name + ". Wails shell ready."
}