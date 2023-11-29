# usage
```go
nt := NewThread(4, 64)
go func() {
    for i := 0; i < 9999; i++ {
        i := i
        nt.AddWork(func(ctx context.Context) {
        //time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
            fmt.Println(i, ctx.Value("n"))
        })
    }
}()

go func() {
    time.Sleep(time.Second * 2)
    nt.Stop()
}()

nt.Wait()
```