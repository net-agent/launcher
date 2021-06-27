package main

import (
	"fmt"
	"math/rand"
	"os"
	"path"
)

func GenerateIPCPath(tempPath string) (string, error) {

	os.Mkdir(tempPath, 0666)

	pid := os.Getpid()
	rnd := rand.Int31()
	fileName := fmt.Sprintf("file-%v-%v.ipc", pid, rnd)

	filePath := path.Join(tempPath, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	file.Close()

	return filePath, nil
}
