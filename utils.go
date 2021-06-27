package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

func ReadJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func WriteJSON(w http.ResponseWriter, errMsg error, data interface{}) {
	resp := &struct {
		ErrCode int
		ErrMsg  string
		Data    interface{}
	}{}

	if errMsg != nil {
		resp.ErrCode = -1
		resp.ErrMsg = errMsg.Error()
	} else {
		resp.ErrCode = 0
		resp.ErrMsg = ""
		resp.Data = data
	}

	buf, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("json marshal failed"))
		return
	}
	w.Write(buf)
}

func ParseRespJSON(r io.Reader, v interface{}) error {
	var resp struct {
		ErrCode int
		ErrMsg  string
		Data    interface{}
	}
	resp.Data = v

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buf, &resp)
	if err != nil {
		return err
	}
	if resp.ErrCode != 0 {
		return errors.New(resp.ErrMsg)
	}
	return nil
}

// LoadJSONFile 加载json文件到对象里
func LoadJSONFile(pathname string, v interface{}) error {
	buf, err := ioutil.ReadFile(pathname)
	if err != nil {
		return err
	}

	// 去掉双斜杠注释
	re := regexp.MustCompile(`(^|\n)\s*\/\/.*`)
	jsonBuf := re.ReplaceAll(buf, nil)

	return json.Unmarshal(jsonBuf, v)
}

// GetLatestBin 按照<prefix><ver><ext>规则，查找ver值最高的文件路径
// 例如：agent_windows_amd64_1234.exe
//   prefix: agent_windows_amd64_
//      ver: 1234
//      ext: .exe
func GetLatestBin(root, prefix, ext string) (binName string, ver int64, err error) {
	s := fmt.Sprintf("^%v\\d+[a-zA-Z]%v$", prefix, ext)
	re := regexp.MustCompile(s)

	latestVer := int64(-1)
	binName = ""

	err = filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := fi.Name()

		if !re.MatchString(name) {
			return nil
		}

		verStr := name[len(prefix) : len(name)-len(ext)]
		ver, err := strconv.ParseInt(verStr, 36, 64)
		if err != nil {
			return nil
		}

		if ver > latestVer {
			latestVer = ver
			binName = name
		}

		return nil
	})

	if latestVer == -1 {
		return "", -1, fmt.Errorf("not found, prefix=%v, ext=%v", prefix, ext)
	}

	return binName, latestVer, err
}
