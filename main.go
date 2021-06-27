package main

import (
	"log"
	"math/rand"
	"os"
	"time"
)

var (
	lg = log.New(os.Stdout, "[launcher]", log.Ldate|log.Ltime|log.Lshortfile)
)

func init() {
	// 初始化随机数
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var err error
	var flags AgentFlags
	flags.Parse()

	err = flags.InitEnv()
	if err != nil {
		lg.Fatal("init env failed:", err)
	}

	var cfg Config
	err = LoadJSONFile(flags.ConfigFileName, &cfg)
	if err != nil {
		lg.Fatal("load config failed:", err)
	}

	switch cfg.Launcher.Mode {
	case "client":
		lg.Printf("run in client mode. addr=%v\n", cfg.Launcher.Address)
		mainClient(&cfg)
	case "server":
		lg.Printf("run in server mode. addr=%v\n", cfg.Launcher.Address)
		mainServer(&cfg)
	default:
		lg.Printf("unknown mode=%v\n", cfg.Launcher.Mode)
	}
}
