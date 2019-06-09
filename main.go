package main

import (
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/lpicanco/i3-autodisplay/config"
	"log"
)

func main() {
	fmt.Println(config.Config)

	X, err := xgb.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	defer X.Close()

	err = randr.Init(X)
	if err != nil {
		log.Fatal(err)
	}

	root := xproto.Setup(X).DefaultScreen(X).Root

	resources, err := randr.GetScreenResources(X, root).Reply()
	if err != nil {
		log.Fatal(err)
	}

	for _, output := range resources.Outputs {
		info, err := randr.GetOutputInfo(X, output, 0).Reply()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s, connected: %t, status: %s \n", info.Name, info.Connection == randr.ConnectionConnected, info)
	}

	err = randr.SelectInputChecked(X, root,
		randr.NotifyMaskScreenChange|randr.NotifyMaskCrtcChange|randr.NotifyMaskOutputChange).Check()
	if err != nil {
		log.Fatal(err)
	}

	for {
		ev, err := X.WaitForEvent()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(ev)

		switch ev.(type) {
		case randr.ScreenChangeNotifyEvent:
			fmt.Println(ev)
		}
	}
}
