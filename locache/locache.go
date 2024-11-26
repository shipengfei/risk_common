package locache

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// 本地缓存

// bool == true, 清理缓存
type cacheWatcher func(*localCache, time.Time) (int64, bool)

type CacheOpt struct {
	KeepAlive    int64 // 精确到秒
	CacheWatcher cacheWatcher
}

type localCache struct {
	*CacheOpt
	*sync.Map
	cycle int64
}

var defaultCacheOpt = &CacheOpt{
	KeepAlive: 60,
	CacheWatcher: func(lc *localCache, t time.Time) (int64, bool) {
		if lc.KeepAlive <= 0 {
			return 0, false
		}
		curCycle := t.Unix() / lc.KeepAlive
		return curCycle, curCycle != atomic.LoadInt64(&lc.cycle)
	},
}

type CacheOpts func(*CacheOpt)

func WithKeepAlive(second int64) CacheOpts {
	return func(co *CacheOpt) {
		co.KeepAlive = second
	}
}

func WithCacheWatcher(watcher cacheWatcher) CacheOpts {
	return func(co *CacheOpt) {
		co.CacheWatcher = watcher
	}
}

func NewLocalCache(opts ...CacheOpts) *localCache {
	cache := &localCache{CacheOpt: defaultCacheOpt}

	for idx := range opts {
		opts[idx](cache.CacheOpt)
	}

	cache.reset()
	go cache.watch()

	return cache
}

func (lc *localCache) reset() {
	// lc.Map = &sync.Map{}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&lc.Map)), unsafe.Pointer(new(sync.Map)))
}

// 每秒检查一次 缓存是否过期
func (lc *localCache) watch() {
	tk := time.NewTicker(time.Second)
	for {
		curTime := <-tk.C

		if lc.CacheWatcher == nil {
			continue
		}

		if curCycle, ok := lc.CacheWatcher(lc, curTime); ok {
			lc.reset()
			atomic.StoreInt64(&lc.cycle, curCycle)
		}

	}
}
