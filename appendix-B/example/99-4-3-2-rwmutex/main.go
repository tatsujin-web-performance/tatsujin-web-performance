package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	// 10秒でタイムアウトする context.Context の生成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// int のスライスの生成
	// 空だと goroutine の実行タイミング次第でゼロ除算エラーになるので先に値を1つ入れておく
	responseTimes := []int{200}
	// sync.RWMutex の生成
	responseTimeMutex := &sync.RWMutex{}

	// 処理の待ち合わせのために sync.WaitGroup を生成
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			// responseTimes の書き込みロックを取得
			// この書き込みロックの間は読み取りロックもブロックされる
			responseTimeMutex.Lock()
			// 100 〜 199 の範囲でランダムな数値を生成しスライスへ追加
			responseTimes = append(responseTimes, rand.Intn(100)+100)
			// responseTimes の書き込みロックを解除
			responseTimeMutex.Unlock()

			// context.Context の終了にあわせてループを抜けるか 100 ミリ秒待つ
			select {
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			// responseTimes の読み取りロックを取得
			// このロックは別の読み取りロックとは競合しないが、書き込みロックはブロックされる
			responseTimeMutex.RLock()

			// responseTimes の個数を標準出力へ表示
			fmt.Printf("response times count: %d\n", len(responseTimes))

			// responseTimes の読み取りロックを解除
			responseTimeMutex.RUnlock()

			// context.Context の終了にあわせてループを抜けるか 1 秒待つ
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			// responseTimes の読み取りロックを取得
			// このロックは別の読み取りロックとは競合しないが、書き込みロックはブロックされる
			responseTimeMutex.RLock()

			// responseTimes の合計値を取得後、個数で割って平均値を算出
			responseTimeSum := 0
			responseTimeCount := len(responseTimes)
			for _, responseTime := range responseTimes {
				responseTimeSum += responseTime
			}
			responseTimeAverage := responseTimeSum / responseTimeCount

			// responseTimes の読み取りロックを解除
			responseTimeMutex.RUnlock()

			// 算出した平均値と個数を標準出力に表示
			fmt.Printf("response times average: %d / %d\n", responseTimeAverage, responseTimeCount)

			// context.Context の終了にあわせてループを抜けるか 1 秒待つ
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}()

	// すべての処理が終わるのを待つ
	wg.Wait()
}