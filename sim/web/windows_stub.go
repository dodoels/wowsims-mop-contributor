//go:build !windows

package main

// Stub implementations for non-Windows platforms

func initWindows() bool {
	// Always return true on non-Windows platforms
	return true
}

func hideConsoleWindow() {
	// No-op on non-Windows platforms
}

func showMessageBox(title, message string) {
	// No-op on non-Windows platforms
}

func checkSingleInstance() bool {
	// Always return true on non-Windows platforms (no single instance check)
	return true
}
