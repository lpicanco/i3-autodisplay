package i3

import (
	"encoding/json"
	"fmt"
	"log"
)

// Version describes an i3 version.
//
// See https://i3wm.org/docs/ipc.html#_version_reply for more details.
type Version struct {
	Major                int64  `json:"major"`
	Minor                int64  `json:"minor"`
	Patch                int64  `json:"patch"`
	Variant              string `json:"variant,omitempty"`
	HumanReadable        string `json:"human_readable"`
	LoadedConfigFileName string `json:"loaded_config_file_name"`
}

// GetVersion returns i3’s version.
//
// GetVersion is supported in i3 ≥ v4.3 (2012-09-19).
func GetVersion() (Version, error) {
	reply, err := roundTrip(messageTypeGetVersion, nil)
	if err != nil {
		return Version{}, err
	}

	var v Version
	err = json.Unmarshal(reply.Payload, &v)
	return v, err
}

// version is a lazily-initialized, possibly stale copy of i3’s GET_VERSION
// reply. Access only values which don’t change, e.g. Major, Minor.
var version Version

// AtLeast returns nil if i3’s major version matches major and i3’s minor
// version is at least minor or newer. Otherwise, it returns an error message
// stating i3 is too old.
func AtLeast(major int64, minor int64) error {
	if major == 0 {
		return fmt.Errorf("BUG: major == 0 is non-sensical. Is a lookup table entry missing?")
	}
	if version.Major == 0 {
		var err error
		version, err = GetVersion()
		if err != nil {
			return err
		}
	}
	if version.Variant != "" {
		log.Printf("non standard i3 payload variant '%s' detected. Ignoring version check. This is fully unsupported.", version.Variant)
		return nil
	}
	if version.Major == major && version.Minor >= minor {
		return nil
	}

	return fmt.Errorf("i3 version too old: got %d.%d, want ≥ %d.%d", version.Major, version.Minor, major, minor)
}
