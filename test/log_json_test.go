package test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"gitlab.miliantech.com/go/common/util/member_crypt"
)

func TestReadCodetagLogFile(t *testing.T) {
	lines, _ := readFileLines("/Users/shipengfei/Downloads/1ab243e7-dd23-403b-a2de-cf7f86cdf0fb.json")
	// t.Log(lines)

	type CodeTag struct {
		CodeTag string
		C       string `json:"c"`
	}

	records := make([]CodeTag, 0)
	for _, line := range lines {
		if line == "" {
			continue
		}
		record := CodeTag{}
		err := json.Unmarshal([]byte(line), &record)
		if err != nil {
			t.Error(err)
		}
		records = append(records, record)
	}
	t.Log(records)
}

func TestParseUids(t *testing.T) {
	lines, _ := readFileLines("/Users/shipengfei/Desktop/member_uid.txt")
	ids, fullOK, err := parseMemberIds(lines)
	if err != nil {
		t.Error(err)
	}
	if !fullOK {
		t.Log("!fullOK")
	}
	t.Log(ids)
}

func TestReadFile(t *testing.T) {
	t.Log([]byte("\n"))
	bs, err := ioutil.ReadFile("/Users/shipengfei/Desktop/member_uid.txt")
	if err != nil {
		t.Fatal(err)
	}

	for idx := bytes.IndexByte(bs, 10); idx > 0; {
		t.Log(string(bs[:idx]))
		if idx == len(bs) { // 说明内容读取完毕
			break
		}
		bs = bs[idx+1:]
	}
	// t.Log(bs)
}

// 批量解析用户Id
func parseMemberIds(values []string) (items []int64, fullOk bool, err error) {
	fullOk = true
	for _, value := range values {
		if value == "" {
			continue
		}
		id, e := member_crypt.Decrypt(value)
		if err = e; err != nil {
			fullOk = false
			return
		}
		items = append(items, int64(id))
	}
	return
}

// 读取文件的内容, 按照行内容返回
func readFileLines(filename string) (lines []string, err error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	lines = strings.Split(string(content), "\n")
	return
}
