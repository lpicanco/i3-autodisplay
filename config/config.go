package config

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type Display struct {
	Name              string
	RandrExtraOptions string   `yaml:"randr_extra_options"`
	TurnOffWhen       []string `yaml:"turn_off_when"`
	Workspaces        []int
}

var Config = struct {
	Displays []Display
}{}

func init() {
	configFile := getConfirFilePath()
	data, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Fatalf("error reading configuration file: %s", err)
	}

	if err := yaml.Unmarshal(data, &Config); err != nil {
		log.Fatalf("error processing configuration file %s: \n %s", configFile, err)
	}
}

func getConfirFilePath() (configFile string) {
	configDir := os.Getenv("XDG_HOME")
	if configDir == "" {
		configDir = path.Join(os.Getenv("HOME"), ".config")
	}

	flag.StringVar(&configFile, "config", path.Join(configDir, "i3-autodisplay", "config.yml"), "Path to configuration file.")
	flag.Parse()

	return
}
