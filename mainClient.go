package main

import (
	"os"
	"time"
)

func mainClient(cfg *Config) {

	// 生成临时文件
	ipcPath, err := GenerateIPCPath("./temp")
	if err != nil {
		lg.Fatal("create temp file failed:", err)
	}
	defer os.Remove(ipcPath)

	bi := NewBinInfo(
		"agent",
		"./bin",
		[]string{cfg.GetServerAPI()})

	//
	// 开始循环调用可执行文件
	//
	for {
		lg.Println("check update...")

		errCheckRemote := bi.GetRemoteVersion()
		if errCheckRemote != nil {
			lg.Printf("check update failed: %v\n", errCheckRemote)
		}

		errGetLocal := bi.GetLocalVersion()
		if errGetLocal != nil {
			lg.Printf("get local bin failed: %v\n", errGetLocal)
		}

		// 检测本地和远程都失败，停止运行
		if errCheckRemote != nil && errGetLocal != nil {
			lg.Printf("check version failed, stop launcher")
			return
		}

		if errGetLocal == nil && bi.LocalVersion.Version >= bi.RemoteVersion.Version {
			lg.Printf("your local bin has been updated to latest\n")
		} else {
			err = bi.DownloadToLocal()
			if err != nil {
				lg.Printf("download latest version failed: %v\n", err)
				if errGetLocal != nil {
					lg.Fatal("can't find bin file in local")
				} else {
					lg.Println("update failed, using local version")
				}
			} else {
				lg.Println("update success")
			}
		}

		err = bi.Exec()
		if err != nil {
			lg.Printf("command exec failed: %v\n", err)
		}

		lg.Println("reload agent after 3s...")
		<-time.After(time.Second * 3)
	}
}
