package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
)

var Config = struct {
	Displays []struct {
		Name              string `required:"true"`
		RandrExtraOptions string `yaml:"randr_extra_options"`
		Workspaces        []int
	}
}{}

func init() {
	configFile := getConfirFilePath()
	data, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Fatalf("Error reading configuration file: %s", err)
	}

	if err := yaml.Unmarshal(data, &Config); err != nil {
		log.Fatalf("Error processing configuration file %s: \n %s", configFile, err)
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
