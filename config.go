package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

// Config holds the configuration options for the application.
//
// At the moment, it is quite limited, only supporting the home folder and the
// file name of the metadata.
type Config struct {
	Home string `env:"NAP_HOME" yaml:"home"`
	File string `env:"NAP_FILE" yaml:"file"`

	DefaultLanguage string `env:"NAP_DEFAULT_LANGUAGE" yaml:"default_language"`

	Theme string `env:"NAP_THEME" yaml:"theme"`

	PrimaryColor        string `env:"NAP_PRIMARY_COLOR" yaml:"primary_color"`
	PrimaryColorSubdued string `env:"NAP_PRIMARY_COLOR_SUBDUED" yaml:"primary_color_subdued"`
	BrightGreenColor    string `env:"NAP_BRIGHT_GREEN" yaml:"bright_green"`
	GreenColor          string `env:"NAP_GREEN" yaml:"green"`
	BrightRedColor      string `env:"NAP_BRIGHT_RED" yaml:"bright_red"`
	RedColor            string `env:"NAP_RED" yaml:"red"`
	GrayColor           string `env:"NAP_GRAY" yaml:"gray"`
	TextColor           string `env:"NAP_TEXT" yaml:"foreground"`
	TextInvertColor     string `env:"NAP_TEXTINVERT" yaml:"textinvert"`
	SubTextColor        string `env:"NAP_SUBTEXT" yaml:"subtext"`
}

func newConfig() Config {
	return Config{
		Home:                defaultHome(),
		File:                "snippets.json",
		DefaultLanguage:     defaultLanguage,
		Theme:               "catppuccin-mocha",
		PrimaryColor:        "#74c7ec",
		PrimaryColorSubdued: "#94e2d5",
		BrightGreenColor:    "#f9e2af",
		GreenColor:          "#a6e3a1",
		BrightRedColor:      "#eba0ac",
		RedColor:            "#f38ba8",
		GrayColor:           "#313244",
		TextColor:           "#cdd6f4",
		SubTextColor:        "#6c7086",
		TextInvertColor:     "#11111b",
	}
}

// default helpers for the configuration.
// We use $XDG_DATA_HOME to avoid cluttering the user's home directory.
func defaultHome() string { return filepath.Join(xdg.DataHome, "nap") }

// defaultConfig returns the default config path
func defaultConfig() string {
	if c := os.Getenv("NAP_CONFIG"); c != "" {
		return c
	}
	cfgPath, err := xdg.ConfigFile("nap/config.yaml")
	if err != nil {
		return "config.yaml"
	}
	return cfgPath
}

// readConfig returns a configuration read from the environment.
func readConfig() Config {
	config := newConfig()
	fi, err := os.Open(defaultConfig())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return newConfig()
	}
	if fi != nil {
		defer fi.Close()
		if err := yaml.NewDecoder(fi).Decode(&config); err != nil {
			return newConfig()
		}
	}

	if err := env.Parse(&config); err != nil {
		return newConfig()
	}

	if strings.HasPrefix(config.Home, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			config.Home = filepath.Join(home, config.Home[1:])
		}
	}

	return config
}

// writeConfig returns a configuration read from the environment.
func (config Config) writeConfig() error {
	fi, err := os.Create(defaultConfig())
	if err != nil {
		return err
	}
	if fi != nil {
		defer fi.Close()
		if err := yaml.NewEncoder(fi).Encode(&config); err != nil {
			return err
		}
	}

	return nil
}
