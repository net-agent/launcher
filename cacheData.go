package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type cacheData struct {
	Path        string
	Name        string
	UpdatedTime time.Time
	Data        []byte
	rwmut       sync.RWMutex
}

func (cd *cacheData) Update() error {
	cd.rwmut.Lock()
	defer cd.rwmut.Unlock()

	data, err := ioutil.ReadFile(cd.Path)
	if err != nil {
		return err
	}

	cd.Data = data
	cd.UpdatedTime = time.Now().Add(time.Minute * 10) // 缓存有效期，10分钟

	return nil
}

func (cd *cacheData) Download(w http.ResponseWriter) {
	if cd.Expired() {
		err := cd.Update()
		if err != nil {
			WriteJSON(w, err, nil)
			return
		}
	}
	cd.rwmut.RLock()
	buf := cd.Data
	cd.rwmut.RUnlock()
	if buf == nil {
		WriteJSON(w, errors.New("data is nil"), nil)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v", cd.Name))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%v", len(buf)))
	w.WriteHeader(http.StatusOK)

	// 限制下载流速500KB/s，每秒10次，每次50KB
	ticker := time.NewTicker(time.Millisecond * 50)
	defer ticker.Stop()
	posStart := 0
	size := 25 * 1024
	for posStart < len(buf) {
		<-ticker.C
		posEnd := posStart + size
		if posEnd > len(buf) {
			posEnd = len(buf)
		}
		_, err := w.Write(buf[posStart:posEnd])
		if err != nil {
			return
		}
		posStart = posEnd
	}
}

func (cd *cacheData) Expired() bool {
	cd.rwmut.RLock()
	defer cd.rwmut.RUnlock()

	return cd.UpdatedTime.Before(time.Now())
}
