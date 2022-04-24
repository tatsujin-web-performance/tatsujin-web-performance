package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

var group singleflight.Group

type Cache struct {
	mu    sync.Mutex
	items map[int]int
}

func NewCache() *Cache {
	m := make(map[int]int)
	c := &Cache{
		items: m,
	}
	return c
}

func (c *Cache) Set(key int, value int) {
	c.mu.Lock()
	c.items[key] = value
	c.mu.Unlock()
}

func (c *Cache) Get(key int) int {
	c.mu.Lock()
	v, ok := c.items[key]
	c.mu.Unlock()

	if ok {
		return v
	}

	// singleflightを使うと複数回同時に呼び出された場合は2つ目以降は1つ目の実行が終了するのを待つ
	vv, err, _ := group.Do(fmt.Sprintf("cacheGet_%d", key), func() (interface{}, error) {
		value := HeavyGet(key)
		c.Set(key, value)
		return value, nil
	})

	if err != nil {
		panic(err)
	}

	// interface{}型なのでint型にキャスト
	return vv.(int)
}

// 実際にはデータベースへのアクセスなどが発生する
// 今回は仮に1秒sleepしてからkeyの2倍を返す
func HeavyGet(key int) int {
	log.Printf("call HeavyGet %d\n", key)
	time.Sleep(time.Second)
	return key * 2
}

func main() {
	mCache := NewCache()

	for i := 0; i < 100; i++ {
		go func(i int) {
			// 0から9までの各キーをほぼ同時に10回取得するがそれぞれ一度しかHeavyGetは実行されない
			mCache.Get(i % 10)
		}(i)
	}

	time.Sleep(2 * time.Second)

	for i := 0; i < 10; i++ {
		log.Println(mCache.Get(i))
	}
}
