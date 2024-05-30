package duration_cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type TestDCache struct {
	A int    `json:"a"`
	B string `json:"b"`
}

func TestGet(t *testing.T) {
	var (
		n   = 10
		wg  = sync.WaitGroup{}
		key = "test_key"
	)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rs := &TestDCache{}
			err := Get(key, func() (*TestDCache, error) {
				fmt.Println("hit once")
				return &TestDCache{
					A: 123,
					B: "123",
				}, nil
			}, 100, &rs)
			if err != nil {
				panic(err.Error())
			}
			fmt.Printf("%v\n", rs)
		}()
	}
	time.Sleep(1 * time.Second)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rs := &TestDCache{}
			err := Get(key, func() (*TestDCache, error) {
				fmt.Println("hit once")
				return &TestDCache{
					A: 123,
					B: "123",
				}, nil
			}, 100, &rs)
			if err != nil {
				panic(err.Error())
			}
			fmt.Printf("cache %v\n", rs)
		}()
	}
	wg.Wait()
	c, _ := gCache.Get([]byte(key))
	fmt.Printf("%s\n", string(c))
}
