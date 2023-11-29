package fast_thread

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewThread(t *testing.T) {
	nt := NewThread(4, 64)
	go func() {
		for i := 0; i < 9999; i++ {
			i := i
			nt.AddWork(func(ctx context.Context) {
				//time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
				fmt.Println(i, ctx.Value("n"))

			})
		}
		time.Sleep(time.Second * 5)
		for i := 9999; i < 19999; i++ {
			i := i
			nt.AddWork(func(ctx context.Context) {
				fmt.Println(i, ctx.Value("n"))
			})
		}
	}()
	go func() {
		time.Sleep(time.Second * 2)
		nt.Stop()
	}()

	nt.Wait()
}
