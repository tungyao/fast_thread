package fast_thread

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Thread struct {
	n           int64
	liveN       int64
	queue       chan func(context.Context)
	t           []*work
	cancel      context.CancelFunc
	ctx         context.Context
	ready       bool
	buffLen     int
	fix         int64
	rateLimiter *time.Ticker // Rate limiter for controlling concurrency
}

type work struct {
	no     int
	status int64
	queue  chan func(ctx context.Context)
	ctx    context.Context
}

// NewThread creates a new thread with specified number of threads and channel buffer length
func NewThread(n int64, chanBuffLen int) *Thread {
	t := &Thread{
		n:           n,
		liveN:       0,
		t:           make([]*work, int(n)),
		ctx:         context.Background(),
		buffLen:     chanBuffLen,
		ready:       true,
		rateLimiter: time.NewTicker(time.Second), // Initialize rate limiter
	}

	if t.buffLen == 0 {
		t.buffLen = 64
	}

	sg := &sync.WaitGroup{}
	sg.Add(int(n))
	for i := 0; i < int(n); i++ {
		go t.Work(i, sg)
	}
	sg.Wait()

	go func() {
		for {
			if atomic.LoadInt64(&t.liveN) == n {
				log.Println(t.liveN, n)
				break
			}
		}
	}()

	return t
}

// SetRateLimit sets the rate limit for the thread
func (t *Thread) SetRateLimit(limit int) {
	t.rateLimiter = time.NewTicker(time.Second / time.Duration(limit))
}

func (t *Thread) Wait() {
	ctx, cancel := context.WithCancel(t.ctx)
	t.cancel = cancel
	<-ctx.Done()
	t.ready = false
	for i := 0; i < int(t.n); i++ {
		close(t.t[i].queue)
	}
}

func NewThreadFixation(n int64, chanBuffLen int, fixation int64) *Thread {
	t := &Thread{
		n:       n,
		liveN:   0,
		t:       make([]*work, int(n)),
		ctx:     context.Background(),
		buffLen: chanBuffLen,
		fix:     0,
		ready:   true,
	}
	if t.buffLen == 0 {
		t.buffLen = 64
	}
	// init thread
	sg := &sync.WaitGroup{}
	sg.Add(int(n))
	for i := 0; i < int(n); i++ {
		go t.Work(i, sg)
	}
	sg.Wait()
	go func(t *Thread) {
		for {
			if atomic.LoadInt64(&t.fix) == fixation {
				t.Stop()
				break
			}
		}
	}(t)
	return t
}

// WaitFixation 等待确定个数的执行完就立即终止
func (t *Thread) WaitFixation() {
	ctx, cancel := context.WithCancel(t.ctx)
	t.cancel = cancel
	<-ctx.Done()
	t.ready = false
	for i := 0; i < int(t.n); i++ {
		close(t.t[i].queue)
	}
	if t.rateLimiter != nil {
		t.rateLimiter.Stop()
	}
}
func (t *Thread) AddWork(fnc func(ctx context.Context)) {
	if t.ready == false {
		return
	}
	m := atomic.LoadInt64(&t.t[0].status)
	no := 0
	for i, w := range t.t {
		n := atomic.LoadInt64(&w.status)
		if n < m {
			m = n
			no = i
		}
	}
	t.t[no].queue <- fnc
	atomic.AddInt64(&t.t[no].status, 1)
}
func (t *Thread) Stop() {
	t.cancel()
}

func (t *Thread) Work(n int, sg *sync.WaitGroup) {
	w := &work{
		no:    n,
		queue: make(chan func(ctx context.Context), t.buffLen),
		ctx:   context.Background(),
	}
	ctx := context.WithValue(w.ctx, "n", n)
	t.t[n] = w
	sg.Done()
	atomic.AddInt64(&t.liveN, 1)

	for v := range w.queue {
		if t.rateLimiter != nil {
			<-t.rateLimiter.C // 等待 rateLimiter 的信号
		}
		v(ctx)
		atomic.AddInt64(&w.status, -1)
		atomic.AddInt64(&t.fix, 1)
	}
}
