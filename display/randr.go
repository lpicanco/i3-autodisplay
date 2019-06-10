package display

import (
	"log"
	"os/exec"
	"reflect"
	"strings"

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
		log.Fatal(err)
	}

	err = randr.Init(xgbConn)
	if err != nil {
		log.Fatal(err)
	}
}

func Refresh() {
	currentOutputConfiguration := getOutputConfiguration()

	if reflect.DeepEqual(currentOutputConfiguration, lastOutputConfiguration) {
		return
	}

	for _, display := range config.Config.Displays {
		if currentOutputConfiguration[display.Name] {
			refreshDisplay(display)
		}
	}

	lastOutputConfiguration = currentOutputConfiguration
}

func ListenEvents() {
	defer xgbConn.Close()

	root := xproto.Setup(xgbConn).DefaultScreen(xgbConn).Root
	err := randr.SelectInputChecked(xgbConn, root,
		randr.NotifyMaskScreenChange|randr.NotifyMaskCrtcChange|randr.NotifyMaskOutputChange).Check()

	if err != nil {
		log.Fatal(err)
	}

	for {
		ev, err := xgbConn.WaitForEvent()
		if err != nil {
			log.Fatal(err)
		}

		switch ev.(type) {
		case randr.ScreenChangeNotifyEvent:
			Refresh()
		}
	}
}

func refreshDisplay(display config.Display) {
	args := []string{"--output", display.Name, "--auto"}
	if display.RandrExtraOptions != "" {
		args = append(args, strings.Split(display.RandrExtraOptions, " ")...)
	}

	cmd := exec.Command("xrandr", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error executing xrandr: %s\n%s", err, out)
	}
}

func getOutputConfiguration() map[string]bool {
	config := make(map[string]bool)

	root := xproto.Setup(xgbConn).DefaultScreen(xgbConn).Root
	resources, err := randr.GetScreenResources(xgbConn, root).Reply()

	if err != nil {
		log.Fatal(err)
	}

	for _, output := range resources.Outputs {
		info, err := randr.GetOutputInfo(xgbConn, output, 0).Reply()
		if err != nil {
			log.Fatal(err)
		}

		config[string(info.Name)] = info.Connection == randr.ConnectionConnected
	}

	return config
}
