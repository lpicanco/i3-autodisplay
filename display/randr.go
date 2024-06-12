package display

import (
	"log"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/lpicanco/i3-autodisplay/i3"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/lpicanco/i3-autodisplay/config"
)

var (
	xgbConn                 *xgb.Conn
	lastOutputConfiguration map[string]bool
)

func init() {
	var err error
	
	xgbConn, err = xgb.NewConn()
	if err != nil {
		log.Fatalf("error initializing xgb: %v", err)
	}

	err = randr.Init(xgbConn)
	if err != nil {
		log.Fatalf("error initializing randr: %v", err)
	}
}

func Refresh() {
	currentOutputConfiguration := getOutputConfiguration()

	if reflect.DeepEqual(currentOutputConfiguration, lastOutputConfiguration) {
		return
	}

	currentWorkspace, err := i3.GetCurrentWorkspaceNumber()
	if err != nil {
		log.Fatalf("error getting i3 current workspace: %v", err)
	}

	args := []string{}
	for _, display := range config.Config.Displays {
		active := isDisplayActive(display, currentOutputConfiguration)
		args = append(args, getDisplayOptions(display, active)...)
	}

	// sometimes during suspend/wake up routines xrandr is unable to configure monitors.
	// 5 secs of retries should be enough to work it around
	retries := 5
	retriesLeft := retries

	for {
		log.Println("xrandr", args)
		cmd := exec.Command("xrandr", args...)
		out, err := cmd.CombinedOutput()
		if err == nil {
			break
		}
		log.Printf("Error executing xrandr: %s\n%s", err, out)
		retriesLeft--
		if retriesLeft == 0 {
			log.Fatalf("unable to execute the xrandr command after %d attempts", retries)
		}
		time.Sleep(time.Second)
	}

	for _, display := range config.Config.Displays {
		if isDisplayActive(display, currentOutputConfiguration) {
			refreshDisplay(display)
		}
	}

	err = i3.SetCurrentWorkspace(currentWorkspace)
	if err != nil {
		log.Fatalf("error setting i3 current workspace: %v", err)
	}

	lastOutputConfiguration = currentOutputConfiguration
}

func ListenEvents() {
	defer xgbConn.Close()

	root := xproto.Setup(xgbConn).DefaultScreen(xgbConn).Root
	err := randr.SelectInputChecked(xgbConn, root,
		randr.NotifyMaskScreenChange|randr.NotifyMaskCrtcChange|randr.NotifyMaskOutputChange).Check()

	if err != nil {
		log.Fatalf("error subscribing to randr events: %v", err)
	}

	for {
		ev, err := xgbConn.WaitForEvent()
		if err != nil {
			log.Fatalf("error processing randr event: %v", err)
		}

		switch ev.(type) {
		case randr.ScreenChangeNotifyEvent:
			Refresh()
		}
	}
}

func isDisplayActive(display config.Display, currentOutputConfiguration map[string]bool) bool {
	if !currentOutputConfiguration[display.Name] {
		return false
	}
	for _, d := range display.TurnOffWhen {
		if currentOutputConfiguration[d] {
			return false
		}
	}
	return true
}

func getDisplayOptions(display config.Display, active bool) []string {
	if active {
		args := []string{"--output", display.Name, "--auto"}
		if display.RandrExtraOptions != "" {
			args = append(args, strings.Split(display.RandrExtraOptions, " ")...)
		}
		return args
	} else {
		args := []string{"--output", display.Name, "--off"}
		return args
	}
}

func refreshDisplay(display config.Display) {
	err := i3.UpdateWorkspaces(display)
	if err != nil {
		log.Fatalf("Error updating i3 workspaces: %s\n", err)
	}
}

func getOutputConfiguration() map[string]bool {
	config := make(map[string]bool)

	root := xproto.Setup(xgbConn).DefaultScreen(xgbConn).Root
	resources, err := randr.GetScreenResources(xgbConn, root).Reply()

	if err != nil {
		log.Fatalf("error getting randr screen resources: %v", err)
	}

	for _, output := range resources.Outputs {
		info, err := randr.GetOutputInfo(xgbConn, output, 0).Reply()
		if err != nil {
			log.Fatalf("error getting randr output info: %v", err)
		}

		config[string(info.Name)] = info.Connection == randr.ConnectionConnected
	}

	return config
}
