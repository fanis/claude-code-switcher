//go:build windows

package process

import (
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW       = kernel32.NewProc("Process32FirstW")
	procProcess32NextW        = kernel32.NewProc("Process32NextW")
	procOpenProcess           = kernel32.NewProc("OpenProcess")
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
	procCloseHandle           = kernel32.NewProc("CloseHandle")

	psapi                     = syscall.NewLazyDLL("psapi.dll")
	procGetProcessImageFileNameW = psapi.NewProc("GetProcessImageFileNameW")

	ntdll                     = syscall.NewLazyDLL("ntdll.dll")
	procNtQueryInformationProcess = ntdll.NewProc("NtQueryInformationProcess")
)

const (
	TH32CS_SNAPPROCESS = 0x00000002
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ = 0x0010
	MAX_PATH = 260
)

type PROCESSENTRY32W struct {
	Size              uint32
	CntUsage          uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	CntThreads        uint32
	ParentProcessID   uint32
	PriClassBase      int32
	Flags             uint32
	ExeFile           [MAX_PATH]uint16
}

// ClaudeProcess represents a running Claude process
type ClaudeProcess struct {
	PID        uint32
	ExePath    string
	WorkingDir string
}

// FindRunningClaudeProcesses finds all running Claude/node processes that might be Claude Code
func FindRunningClaudeProcesses() []ClaudeProcess {
	var processes []ClaudeProcess

	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot == uintptr(syscall.InvalidHandle) {
		return processes
	}
	defer procCloseHandle.Call(snapshot)

	var entry PROCESSENTRY32W
	entry.Size = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return processes
	}

	for {
		exeName := syscall.UTF16ToString(entry.ExeFile[:])
		exeNameLower := strings.ToLower(exeName)

		// Look for node.exe (Claude Code runs via Node.js)
		if exeNameLower == "node.exe" {
			proc := ClaudeProcess{
				PID: entry.ProcessID,
			}

			// Try to get the full executable path
			if path := getProcessPath(entry.ProcessID); path != "" {
				proc.ExePath = path
			}

			// Try to get working directory
			if workDir := getProcessWorkingDirectory(entry.ProcessID); workDir != "" {
				proc.WorkingDir = workDir
				processes = append(processes, proc)
			}
		}

		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	return processes
}

func getProcessPath(pid uint32) string {
	handle, _, _ := procOpenProcess.Call(PROCESS_QUERY_LIMITED_INFORMATION, 0, uintptr(pid))
	if handle == 0 {
		return ""
	}
	defer procCloseHandle.Call(handle)

	var buf [MAX_PATH * 2]uint16
	size := uint32(len(buf))
	ret, _, _ := procQueryFullProcessImageNameW.Call(handle, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if ret == 0 {
		return ""
	}

	return syscall.UTF16ToString(buf[:size])
}

// getProcessWorkingDirectory attempts to get the working directory of a process
// This is complex on Windows and may not always succeed
func getProcessWorkingDirectory(pid uint32) string {
	handle, _, _ := procOpenProcess.Call(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ, 0, uintptr(pid))
	if handle == 0 {
		return ""
	}
	defer procCloseHandle.Call(handle)

	// Getting the working directory on Windows is complex and requires reading
	// the process's PEB (Process Environment Block). For simplicity, we'll
	// use a heuristic: check if any command line arguments contain project paths

	// For now, return empty - we'll use a different approach
	return ""
}

// IsProjectInUse checks if a Claude process is running in the given project path
func IsProjectInUse(projectPath string, processes []ClaudeProcess) bool {
	normalizedProject := strings.ToLower(filepath.Clean(projectPath))

	for _, proc := range processes {
		if proc.WorkingDir != "" {
			normalizedWorkDir := strings.ToLower(filepath.Clean(proc.WorkingDir))
			if normalizedWorkDir == normalizedProject {
				return true
			}
		}
	}

	return false
}

// GetProcessHWND attempts to find the window handle for a process
// This is used to focus an existing terminal window
func GetProcessHWND(pid uint32) uintptr {
	// This would require EnumWindows - implemented separately
	return 0
}
