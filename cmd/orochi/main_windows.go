//go:build windows

package main

import (
	"os"
	"syscall"
)

func init() {
	// On Windows, allocate a console if running from a non-console context
	// This prevents the app from immediately closing when double-clicked
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procAllocConsole := kernel32.NewProc("AllocConsole")
	procAttachConsole := kernel32.NewProc("AttachConsole")
	procGetConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	
	// Check if we already have a console
	ret, _, _ := procGetConsoleWindow.Call()
	if ret == 0 {
		// No console window, try to attach to parent console first
		r1, _, _ := procAttachConsole.Call(uintptr(uint32(^uint32(0)))) // ATTACH_PARENT_PROCESS = -1
		if r1 == 0 {
			// Failed to attach to parent, allocate new console
			procAllocConsole.Call()
		}
		
		// Reopen stdout, stderr, stdin
		os.Stdout = os.NewFile(uintptr(syscall.Stdout), "/dev/stdout")
		os.Stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")
		os.Stdin = os.NewFile(uintptr(syscall.Stdin), "/dev/stdin")
	}
}

// SetupWindowsConsole ensures proper console behavior on Windows
func SetupWindowsConsole() {
	// This function is called by init(), so it's empty
	// Keeping it for potential future use
}