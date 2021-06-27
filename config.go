package main

import "fmt"

type Config struct {
	Launcher LauncherInfo `json:"launcher"`
}

type LauncherInfo struct {
	Mode    string `json:"mode"` // client / server
	Address string `json:"address"`
	Https   bool   `json:"https"`
	ApiPath string `json:"apiPath"`
}

func (cfg *Config) GetServerAPI() string {
	scheme := "http://"
	if cfg.Launcher.Https {
		scheme = "https://"
	}
	apiPath := "/launcher/"
	if cfg.Launcher.ApiPath != "" {
		apiPath = cfg.Launcher.ApiPath
	}

	return fmt.Sprintf("%v%v%v", scheme, cfg.Launcher.Address, apiPath)
}
