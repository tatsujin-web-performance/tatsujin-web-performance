package main

import (
	"log"
	"sync"
	"time"
)

const defaultValue = 100

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

	go func() {
		// 非同期にキャッシュ更新処理を実行する
		v := HeavyGet(key)

		c.Set(key, v)
	}()

	return defaultValue
}

// 実際にはデータベースへのアクセスなどが発生する
// 今回は仮に1秒sleepしてからkeyの2倍を返す
func HeavyGet(key int) int {
	time.Sleep(time.Second)
	return key * 2
}

func main() {
	mCache := NewCache()
	// 最初はデフォルト値が返る
	log.Println(mCache.Get(3))
	time.Sleep(time.Second)
	// 次は更新されている
	log.Println(mCache.Get(3))
}
