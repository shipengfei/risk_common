package utils

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestMapKeysToList(t *testing.T) {
	var m1 = map[string]int64{"a": 1, "b": 2, "c": 3}
	t.Log("keys =>", MapKeysToList(m1, nil))
	t.Log("values =>", MapValuesToList(m1, nil))

	var m2 = map[int64]bool{1: true, 2: true, 3: false}
	t.Log("keys =>", MapKeysToList(m2, func(key int64) bool {
		return key > 1
	}))
	t.Log("values =>", MapValuesToList(m2, nil))

	var m3 = map[string]bool{"def": true, "abc": true, "xxx": false}
	ks, vs := MapToList(m3, func(key string, val bool) bool { return val })
	t.Log(ks, vs)
}

func TestOs(t *testing.T) {
	result := ArrayInter([]string{"a", "b", "b"}, []string{"a", "b", "b", "c"}, false)
	t.Log(result)
}

func TestQPSLimitor(t *testing.T) {
	var opt = QpsOpt{Key: "string", MaxQPS: 200}
	var opt2 = QpsOpt{Key: "string1", MaxQPS: 200}
	var count int64
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			if limit, _ := QPSLimitor(opt); limit {
				atomic.AddInt64(&count, 1)
			}
			wg.Done()
		}(i)

		wg.Add(1)
		go func(i int) {
			if limit, _ := QPSLimitor(opt2); limit {
				atomic.AddInt64(&count, 1)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	t.Log(atomic.LoadInt64(&count))
}
