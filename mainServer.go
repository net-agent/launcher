package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

func mainServer(cfg *Config) {
	r := mux.NewRouter()

	launcher := r.PathPrefix(cfg.Launcher.ApiPath).Subrouter()
	launcher.Methods("GET").Path("/check-latest-bin").HandlerFunc(CheckLatestBinHandler)
	launcher.Methods("GET").Path("/download-bin").HandlerFunc(GetDownloadBinHandler())

	svr := &http.Server{
		Addr:    cfg.Launcher.Address,
		Handler: r,
	}
	svr.ListenAndServe()
}

func CheckLatestBinHandler(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	prefix := v.Get("prefix")
	ext := v.Get("ext")
	name, ver, err := GetLatestBin("./bin", prefix, ext)
	if err != nil {
		WriteJSON(w, err, nil)
		return
	}

	WriteJSON(w, nil, struct {
		Version  int64
		FullName string
		FullPath string
	}{ver, name, ""})
}

func GetDownloadBinHandler() http.HandlerFunc {
	fileDataCache := make(map[string]*cacheData)
	var fileDataCacheMut sync.Mutex

	return func(w http.ResponseWriter, r *http.Request) {

		v := r.URL.Query()
		binName := v.Get("bin")

		if strings.Contains(binName, "\\") ||
			strings.Contains(binName, "..") ||
			strings.Contains(binName, "/") ||
			strings.Contains(binName, "~") {
			WriteJSON(w, errors.New("invalid bin name"), nil)
			return
		}

		fileDataCacheMut.Lock()
		cache, found := fileDataCache[binName]
		if !found {
			cache = &cacheData{
				Path: fmt.Sprintf("./bin/%v", binName),
				Name: binName,
			}
			fileDataCache[binName] = cache
		}
		fileDataCacheMut.Unlock()

		cache.Download(w)
	}
}
