package utils

import (
	"sync"
	"sync/atomic"
	"time"
)

type QpsOpt struct {
	Key    string
	MaxQPS int64
}

type QPS struct {
	CurrSecond int64
	QPS        int64
}

var localQPS = &sync.Map{} //make(map[string]*QPS)

func QPSLimitor(opt QpsOpt) (limit bool, q int64) {
	qpsTmp, _ := localQPS.LoadOrStore(opt.Key, &QPS{})
	qps := qpsTmp.(*QPS)

	currSec := time.Now().Unix()
	if sec := atomic.LoadInt64(&qps.CurrSecond); sec != currSec {
		atomic.StoreInt64(&qps.QPS, 0)
		atomic.StoreInt64(&qps.CurrSecond, currSec)
		return
	}

	currQps := atomic.AddInt64(&qps.QPS, 1)
	return currQps >= opt.MaxQPS, currQps
}
