package main

import (
	"flag"
	"os"
	"path/filepath"
)

type AgentFlags struct {
	HomePath       string
	ConfigFileName string
	IPCPath        string
}

func (f *AgentFlags) Parse() {
	flag.StringVar(&f.HomePath,
		"home", "", "home path for files")
	flag.StringVar(&f.ConfigFileName,
		"c", "./config.json", "default name of config file")
	flag.StringVar(&f.IPCPath,
		"ipc", "", "ipc path for launcher")
	flag.Parse()
}

func (f *AgentFlags) InitEnv() error {
	var err error

	// set home path
	if f.HomePath != "" {
		err := os.Chdir(f.HomePath)
		if err != nil {
			return err
		}
	}

	f.HomePath, err = filepath.Abs(f.HomePath)
	if err != nil {
		return err
	}

	return nil
}
