# i3-autodisplay
i3-autodisplay is a i3wm display auto-configuration

[![Go Report Card](https://goreportcard.com/badge/github.com/lpicanco/i3-autodisplay)](https://goreportcard.com/report/github.com/lpicanco/i3-autodisplay)

## Installation

### Pre-requisites
[xrandr](https://www.x.org/archive/current/doc/man/man1/xrandr.1.xhtml) program

### Pre built binary
Fetch the [latest release](https://github.com/lpicanco/i3-autodisplay/releases).

### From sources

```bash
git clone https://github.com/lpicanco/i3-autodisplay.git
cd i3-autodisplay
go build ./...
```

## Usage
`i3-autodisplay` requires a configuration file to work. The configuration file can be loaded from these locations:

* `$XDG_HOME/i3-autodisplay/config.yml`
* `$HOME/.config/i3-autodisplay/config.yml`
* Specified via `-config` parameter

In your i3wm configuration add the following line:

```
exec --no-startup-id <path to i3-autodisplay>
```

Usage via command line:
```bash
./i3-autodisplay -config sample_config.yml
```

Sample configuration file:
```yaml
displays:
  - name: eDP1
    workspaces: [1,2,3,4,5,6,7,8,9,0]
  - name: HDMI1
    workspaces: [2,4,6,8]
    randr_extra_options: "--left-of eDP1"
  - name: DP1
    workspaces: [1,3,5,7,9]
    randr_extra_options: "--left-of HDMI1"
```

