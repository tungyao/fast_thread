package fast_thread

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestNewThreadFixation(t *testing.T) {
	nt := NewThreadFixation(4, 64, 100)
	go func() {
		for i := 0; i < 100; i++ {
			ii := i
			nt.AddWork(func(ctx context.Context) {
				//log.Println(ii)
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(300)))
				fmt.Println(ii, ctx.Value("n"))

			})
		}
	}()
	nt.WaitFixation()
}

func TestNewThread(t *testing.T) {
	nt := NewThread(4, 64)
	go func() {
		for i := 0; i < 9999; i++ {
			ii := i
			nt.AddWork(func(ctx context.Context) {
				//log.Println(ii)
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(300)))
				fmt.Println(ii, ctx.Value("n"))

			})
		}
	}()
	go func() {
		time.Sleep(time.Second * 3)
		nt.Stop()
	}()
	nt.Wait()
}
