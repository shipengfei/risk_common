package ydcfg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/go-redis/redis/v8"
	"gitlab.miliantech.com/infrastructure/log"
	"gitlab.miliantech.com/infrastructure/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ============Option start============
type Opt struct {
	Dur         time.Duration
	Ctx         context.Context
	DbCli       *gorm.DB
	RdsCli      *redis.Client
	locTk       *time.Ticker
	locCache    *sync.Map
	CfgCacheKey string
}

var DefaultOpt = NewOpt()

func NewOpt() *Opt {
	return &Opt{
		Dur:         time.Minute * 30,
		Ctx:         trace.Context(context.Background()),
		locCache:    new(sync.Map),
		CfgCacheKey: "rc_risk_configurations",
	}
}

func (record *Opt) Init(opts ...Opts) {
	defOpt := record

	for _, o := range opts {
		o(defOpt)
	}
	defOpt.locTk = time.NewTicker(defOpt.Dur)

	var loadData = loadCfgFromRds
	loadData(defOpt)

	go func() {
		for {
			<-defOpt.locTk.C
			loadData(defOpt)
		}
	}()
}

func (record *Opt) Stop() {
	record.locTk.Stop()
	record.locCache = new(sync.Map)
}

type Opts func(*Opt)

func WithTimeout(dur time.Duration) Opts {
	return func(o *Opt) {
		o.Dur = dur
	}
}

func WithContext(ctx context.Context) Opts {
	return func(o *Opt) {
		o.Ctx = ctx
	}
}

func WithGormCli(db *gorm.DB) Opts {
	return func(o *Opt) {
		o.DbCli = db
	}
}

func WithGoRedisCli(rds *redis.Client) Opts {
	return func(o *Opt) {
		o.RdsCli = rds
	}
}

func WithCfgCacheKey(key string) Opts {
	return func(o *Opt) {
		o.CfgCacheKey = key
	}
}

// ============Option end============

func loadCfgFromRds(o *Opt) {
	ctx, cancel := context.WithTimeout(trace.CopyContextFromIncoming(o.Ctx), time.Second)
	defer cancel()

	values, err := o.RdsCli.HGetAll(ctx, o.CfgCacheKey).Result()
	if err != nil {
		log.Error(ctx, "redis_common:loadCfgFromDb", zap.Error(err))
		return
	}

	cache := new(sync.Map)
	for key, val := range values {
		cache.Store(key, val)
	}
	// log.Info(ctx, "redis_common:loadCfgFromDb", zap.Any("go_text", values))
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&o.locCache)), unsafe.Pointer(cache))
	// o.locCache = cache
}

func GetCfgString(ctx context.Context, key string) (val string, ok bool) {
	defer func() {
		log.Debug(ctx, "getCfgString", zap.Any("go_text", map[string]any{"key": key, "val": val, "ok": ok}))
	}()

	cacheVal, load := DefaultOpt.locCache.Load(key)
	if load {
		if v, ok := cacheVal.(string); ok {
			return v, ok
		}

		return fmt.Sprintf("%v", cacheVal), true
	}

	valRds, err := DefaultOpt.RdsCli.HGet(ctx, DefaultOpt.CfgCacheKey, key).Result()
	if err != nil && err != redis.Nil {
		log.Error(ctx, "redis_common:getCfgString", zap.Error(err))
		return
	}
	return valRds, true
}

func GetCfgMap(ctx context.Context, key string) (vals map[string]any, ok bool) {
	valStr, ok := GetCfgString(ctx, key)
	if !ok {
		return
	}

	vals = map[string]any{}
	err := json.Unmarshal([]byte(valStr), &vals)
	if err != nil {
		log.Error(ctx, "redis_common:getCfgMap", zap.Error(err))
		return
	}

	return vals, true
}

func GetCfgSliceString(ctx context.Context, key string) (vals []string) {
	valStr, ok := GetCfgString(ctx, key)
	if !ok {
		return
	}

	vals = strings.Split(valStr, ",")
	for idx := range vals {
		vals[idx] = strings.TrimSpace(vals[idx])
	}
	return
}

func GetCfgWithDefaultValInt64(ctx context.Context, key string, defVal int64) (val int64) {
	valStr, ok := GetCfgString(ctx, key)
	if !ok {
		return defVal
	}
	if valStr == "" {
		return defVal
	}

	valParse, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		log.Error(ctx, "redis_common:getCfgWithDefaultValInt64", zap.String("go_text", key), zap.Error(err))
		return defVal
	}

	return valParse
}

func GetCfgWithStruct(ctx context.Context, key string, obj any) (ok bool) {
	valString, ok := GetCfgString(ctx, key)
	if !ok {
		return
	}
	err := json.Unmarshal([]byte(valString), obj)
	if err != nil {
		log.Error(ctx, "redis_common:getCfgWithStruct", zap.Error(err))
		return
	}
	return true
}
