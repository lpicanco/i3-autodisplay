package i3

import (
	"sync"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"
)

func i3Pid() int {
	xu, err := xgbutil.NewConn()
	if err != nil {
		return -1 // X session terminated
	}
	defer xu.Conn().Close()
	reply, err := xprop.GetProperty(xu, xu.RootWin(), "I3_PID")
	if err != nil {
		return -1 // I3_PID no longer present (X session replaced?)
	}
	num, err := xprop.PropValNum(reply, err)
	if err != nil {
		return -1
	}
	return int(num)
}

var lastPid struct {
	sync.Mutex
	pid int
}

// IsRunningHook provides a method to override the method which detects if i3 is running or not
var IsRunningHook = func() bool {
	lastPid.Lock()
	defer lastPid.Unlock()
	if !wasRestart || lastPid.pid == 0 {
		lastPid.pid = i3Pid()
	}
	return pidValid(lastPid.pid)
}

func i3Running() bool {
	return IsRunningHook()
}
