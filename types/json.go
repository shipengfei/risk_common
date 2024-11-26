package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

// 使 gorm.v1版本支持JSON序列化
// gorm.v2 版本查看gorm文档https://gorm.io/zh_CN/docs/serializer.html
/*
	type MemberSetting struct {
		MemberId int64
		Setting  JSON `gorm:"serializer:json"`
	}
*/
type JSON map[string]any

func (obj JSON) ToString() string {
	bs, _ := json.Marshal(obj)
	return string(bs)
}

func (obj *JSON) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, obj)
	case string:
		return json.Unmarshal([]byte(src), obj)
	}
	return fmt.Errorf("risk_common: not support type of %s", reflect.TypeOf(src))
}

func (obj JSON) Value() (val driver.Value, err error) {
	bs, errBs := json.Marshal(obj)
	return string(bs), errBs
}

func (obj JSON) Merge(vals JSON) JSON {
	if len(vals) == 0 {
		return obj
	}
	for k, v := range vals {
		obj[k] = v
	}
	return obj
}

func (obj JSON) GetInt64(key string) (val int64, err error) {
	// 如果不存在, 则直接返回
	v, load := obj[key]
	if !load {
		return
	}

	return InterfaceToInt64(v)
}

func (obj JSON) GetString(key string) (val string, err error) {
	return InterfaceToStr(obj[key])
}
