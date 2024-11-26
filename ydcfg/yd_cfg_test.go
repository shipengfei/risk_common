package ydcfg

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestCfg(t *testing.T) {
	str, _ := GetCfgMap(context.Background(), "abc")
	t.Log(str)
}

func init() {
	DefaultOpt.
		Init(
			WithTimeout(time.Second),
			WithGoRedisCli(redis.NewClient(&redis.Options{})),
		)
}

func TestGetCfgSliceString(t *testing.T) {
	vals := GetCfgSliceString(context.Background(), "abc")
	t.Log(vals)
}
