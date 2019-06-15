package i3

import "syscall"

func pidValid(pid int) bool {
	// As per kill(2) from POSIX.1-2008, sending signal 0 validates a pid.
	if err := syscall.Kill(pid, 0); err != nil {
		if err == syscall.EPERM {
			// Process still alive (but no permission to signal):
			return true
		}
		// errno is likely ESRCH (process not found).
		return false // Process not alive.
	}
	return true // Process still alive.
}
