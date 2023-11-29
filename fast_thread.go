package fast_thread

import (
	"context"
	"sync/atomic"
)

type Thread struct {
	n       int64                      // total number thread
	liveN   int64                      // running  thread
	queue   chan func(context.Context) // word queue
	t       []*work                    // thread origin
	cancel  context.CancelFunc
	ctx     context.Context
	ready   bool
	buffLen int
}
type work struct {
	no     int
	status int64
	queue  chan func(ctx context.Context)
	ctx    context.Context
}
type burden struct {
}

// NewThread
// n means that set how many thread to start up
func NewThread(n int64, chanBuffLen int) *Thread {
	t := &Thread{
		n:       n,
		liveN:   0,
		t:       make([]*work, int(n)),
		ctx:     context.Background(),
		buffLen: chanBuffLen,
	}
	if t.buffLen == 0 {
		t.buffLen = 64
	}
	// init thread
	for i := 0; i < int(n); i++ {
		go t.Work(i)
	}
	for {
		if atomic.LoadInt64(&t.liveN) == n {
			break
		}
	}
	return t
}

func (t *Thread) Wait() {
	ctx, cancel := context.WithCancel(t.ctx)
	t.cancel = cancel
	<-ctx.Done()
}
func (t *Thread) AddWork(fnc func(ctx context.Context)) {
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

func (t *Thread) Work(n int) {
	w := &work{
		no:    n,
		queue: make(chan func(ctx context.Context), t.buffLen),
		ctx:   context.Background(),
	}
	ctx := context.WithValue(w.ctx, "n", n)
	t.t[n] = w
	atomic.AddInt64(&t.liveN, 1)
	for v := range w.queue {
		v(ctx)
		atomic.AddInt64(&w.status, -1)
	}
}
