package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

type BinVersion struct {
	Version  int64
	FullName string
	FullPath string
}

type BinInfo struct {
	Prefix        string
	Ext           string
	RemoteAPIs    []string
	RemoteVersion BinVersion
	LocalRoot     string
	LocalVersion  BinVersion
}

func NewBinInfo(prefix string, localRoot string, remoteAPIs []string) *BinInfo {

	os.Mkdir(localRoot, 0666)

	ext := ".exe"
	if runtime.GOOS != "windows" {
		ext = "_bin"
	}

	return &BinInfo{
		Prefix:     fmt.Sprintf("%v_%v_", prefix, runtime.GOOS),
		Ext:        ext,
		RemoteAPIs: remoteAPIs,
		LocalRoot:  localRoot,
	}
}

// GetRemoteVersion 从API列表中，选择一个可用API检查服务端版本信息
func (b *BinInfo) GetRemoteVersion() error {
	var resp *http.Response
	var err error = errors.New("empty api list")
	var usingAPI string

	// 从API列表中，选取一个成功连接的地址
	for _, api := range b.RemoteAPIs {
		fullURL := fmt.Sprintf("%v/check-latest-bin?prefix=%v&ext=%v",
			api, url.QueryEscape(b.Prefix), url.QueryEscape(b.Ext))

		resp, err = http.Get(fullURL)
		if err == nil {
			usingAPI = api
			lg.Printf("connect to %v\n", fullURL)
			break
		}
	}

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = ParseRespJSON(resp.Body, &b.RemoteVersion)
	if err != nil {
		return err
	}
	b.RemoteVersion.FullPath = fmt.Sprintf("%v/download-bin?bin=%v",
		usingAPI, b.RemoteVersion.FullName)

	return nil
}

// GetLocalVersion 获取本机环境中的版本信息
func (b *BinInfo) GetLocalVersion() error {
	binName, version, err := GetLatestBin(b.LocalRoot, b.Prefix, b.Ext)
	if err != nil {
		return err
	}
	b.LocalVersion.FullName = binName
	b.LocalVersion.FullPath = path.Join(b.LocalRoot, binName)
	b.LocalVersion.Version = version
	return nil
}

// DownloadToLocal 从云端下载最新版本可执行文件到本地
func (b *BinInfo) DownloadToLocal() error {
	if b.RemoteVersion.Version <= 0 {
		err := b.GetRemoteVersion()
		if err != nil {
			return err
		}
		if b.RemoteVersion.Version <= 0 {
			return errors.New("invalid remote version")
		}
	}

	localFullPath := path.Join(b.LocalRoot, b.RemoteVersion.FullName)
	err := download(b.RemoteVersion.FullPath, localFullPath)
	if err != nil {
		return err
	}
	b.LocalVersion.FullName = b.RemoteVersion.FullName
	b.LocalVersion.FullPath = localFullPath
	b.LocalVersion.Version = b.RemoteVersion.Version
	return nil
}

// Exec 执行本地文件
func (b *BinInfo) Exec(attachArgs ...string) error {
	args := append(os.Args[1:], attachArgs...)
	cmd := exec.Command(b.LocalVersion.FullPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func download(url, localFile string) error {
	p, err := exec.LookPath("wget")
	if err != nil {
		lg.Println("wget not found, use build-in download")
		return goWget(url, localFile)
	}
	lg.Printf("wget found: %v\n", p)
	return wget(url, localFile)
}

func wget(url, localFile string) error {
	cmd := exec.Command("wget", url, "-O", localFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func goWget(url, localFile string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if strings.Contains(resp.Header.Get("Content-Disposition"), "attachment") {
		file, err := os.Create(localFile)
		if err != nil {
			return err
		}
		defer file.Close()

		lg.Println("downloading ...")
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return err
		}
		return nil
	}

	err = ParseRespJSON(resp.Body, nil)
	if err != nil {
		return err
	}

	return errors.New("download failed")
}
