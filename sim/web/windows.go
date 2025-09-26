//go:build windows

package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func hideConsoleWindow() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	showWindow := user32.NewProc("ShowWindow")

	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd != 0 {
		showWindow.Call(hwnd, 0) // SW_HIDE = 0
	}
}

func showMessageBox(title, message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)

	messageBox.Call(0, uintptr(unsafe.Pointer(messagePtr)), uintptr(unsafe.Pointer(titlePtr)), 0)
}

func initWindows() bool {
	if !checkSingleInstance() {
		return false
	}

	hideConsoleWindow()
	return true
}

func checkSingleInstance() bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq wowsimmop-windows.exe", "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		return true
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	runningInstances := 0
	currentPID := os.Getpid()

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) >= 2 {
			pidStr := strings.Trim(fields[1], "\"")
			if pid, err := strconv.Atoi(pidStr); err == nil && pid != currentPID {
				runningInstances++
			}
		}
	}

	if runningInstances > 0 {
		// Silently exit if another instance is already running
		return false
	}

	return true
}
