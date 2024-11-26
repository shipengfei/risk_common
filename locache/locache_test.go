package locache

import (
	"context"
	"testing"
	"time"
)

func TestLocalCache(t *testing.T) {
	cache := NewLocalCache(WithKeepAlive(6))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cache.Store("abc", "123")
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second)
				timeNow := time.Now()
				cache.Store("abc", timeNow.Unix())
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(time.Second)
			t.Log(cache.Load("abc"))
		}
	}
}
