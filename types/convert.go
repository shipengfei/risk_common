package types

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

func StrToInt64(str string) (val int64, err error) {
	return strconv.ParseInt(str, 10, 64)
}

func Int64ToStr(i int64) string {
	return strconv.FormatInt(i, 10)
}

func InterfaceToInt64(src any) (val int64, err error) {
	switch src := src.(type) {
	case int:
		return int64(src), nil
	case int8:
		return int64(src), nil
	case int16:
		return int64(src), nil
	case int32:
		return int64(src), nil
	case int64:
		return src, nil
	case uint:
		return int64(src), nil
	case uint8:
		return int64(src), nil
	case uint16:
		return int64(src), nil
	case uint32:
		return int64(src), nil
	case uint64:
		return int64(src), nil
	case float32:
		return int64(src), nil
	case float64:
		return int64(src), nil
	case time.Time:
		return src.UnixNano(), nil
	case bool:
		if src {
			return 1, nil
		}
		return
	}
	return 0, errors.New("risk_common:not support")
}

func InterfaceToStr(src any) (val string, err error) {
	switch src := src.(type) {
	case string:
		return src, nil
	default:
		return fmt.Sprintf("%v", src), nil
	}
}
