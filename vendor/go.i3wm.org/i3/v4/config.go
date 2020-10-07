package i3

import "encoding/json"

// Config contains details about the configuration file.
//
// See https://i3wm.org/docs/ipc.html#_config_reply for more details.
type Config struct {
	Config string `json:"config"`
}

// GetConfig returns i3’s in-memory copy of the configuration file contents.
//
// GetConfig is supported in i3 ≥ v4.14 (2017-09-04).
func GetConfig() (Config, error) {
	reply, err := roundTrip(messageTypeGetConfig, nil)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(reply.Payload, &cfg)
	return cfg, err
}
