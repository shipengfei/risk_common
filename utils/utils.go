package utils

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"gitlab.miliantech.com/infrastructure/log"
	"go.uber.org/zap"
)

func init() {
	devEnv = strings.Contains(os.Getenv("GOENVMODE"), "test")
}

// 表示测试服环境
var devEnv bool

func IsDevEnv() bool {
	return devEnv
}

func SimpleRecover(ctx context.Context) {
	if err := recover(); err != nil {
		log.Error(ctx, "risk_common.recovered", zap.String("error", string(debug.Stack())))
	}
}

func ArrayContains[K comparable](items []K, val K) bool {
	for _, item := range items {
		if item == val {
			return true
		}
	}
	return false
}

func ArrayInter[K comparable](as, bs []K, uniq bool) (result []K) {
	numCount := make(map[K]int)
	for _, num := range as {
		if uniq {
			numCount[num] = 1
		} else {
			numCount[num]++
		}
	}

	for _, num := range bs {
		if count, ok := numCount[num]; ok && count > 0 {
			result = append(result, num)
			numCount[num]--
		}
	}
	return
}

// Deprecated
//
// return map keys
// if filter is nil, return all keys
func MapKeysToList[K, V comparable](values map[K]V, filter func(K) bool) (items []K) {
	for k := range values {
		if filter != nil && !filter(k) {
			continue
		}
		items = append(items, k)
	}
	return
}

// Deprecated
//
// return map values
// if filter is nil, return all values
func MapValuesToList[K, V comparable](values map[K]V, filter func(V) bool) (items []V) {
	for _, v := range values {
		if filter != nil && !filter(v) {
			continue
		}
		items = append(items, v)
	}
	return
}

func MapToList[K, V comparable](values map[K]V, filter func(K, V) bool) (keys []K, vals []V) {
	for k, v := range values {
		if filter != nil && !filter(k, v) {
			continue
		}
		keys = append(keys, k)
		vals = append(vals, v)
	}
	return
}

func ListToMapBool[Item comparable](items []Item) (vals map[Item]bool) {
	vals = make(map[Item]bool)
	for _, item := range items {
		vals[item] = true
	}
	return
}

func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}

func CamelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if !k && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || !k) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

// 从url中下载文件到本地
// storePath必须是绝对路径
func DownloadSourceFile(remotePath string, storePath string) (err error) {
	client := http.DefaultClient
	client.Timeout = time.Minute
	resp, errResp := client.Get(remotePath)
	if errResp != nil {
		return errResp
	}

	if resp.Body == nil {
		return errors.New("body invalid")
	}
	defer resp.Body.Close()

	file, errCreate := os.Create(storePath)
	if errCreate != nil {
		return errCreate
	}
	defer file.Close()

Break:
	for {
		var buf = make([]byte, 32*1024)
		nRead, errRead := resp.Body.Read(buf)
		if errRead != nil {
			if errRead != io.EOF {
				return errRead
			}
			break Break
		}

		if nRead > 0 {
			nWrite, errWrite := file.Write(buf[0:nRead])
			if errWrite != nil {
				return errWrite
			}

			if nWrite != nRead {
				return errors.New("write size error")
			}
		}
	}
	return
}

func Md5String(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
