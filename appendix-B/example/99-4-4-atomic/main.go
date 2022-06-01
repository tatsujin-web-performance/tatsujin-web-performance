// Windows などの環境では正しく動作しない可能性があります。

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	// 10秒で終了する context.Context の生成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// SIGUSR1 を送るコマンドを例示します
	fmt.Println("Reset counter:")
	fmt.Printf("  $ kill -SIGUSR1 %d\n", os.Getpid())

	// int64 型で0として初期化
	i := int64(0)
	// 最初の値として10を書き込みます
	atomic.StoreInt64(&i, 10)

	// 待ち合わせのために sync.WaitGroup を生成
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			timer := time.After(1 * time.Second)

			// 1秒ごとに現在の値を取得して標準出力に表示します
			n := atomic.LoadInt64(&i)
			fmt.Printf("load now: %d\n", n)

			// 1秒ごとに繰り返すための select
			select {
			case <-ctx.Done():
				return
			case <-timer:
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			timer := time.After(100 * time.Millisecond)

			// 0.1秒ごとにインクリメントします
			// インクリメントした結果の数値を利用したければ返り値を利用してください
			atomic.AddInt64(&i, 1)

			// 0.1秒ごとに繰り返すための select
			select {
			case <-ctx.Done():
				return
			case <-timer:
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			timer := time.After(10 * time.Millisecond)

			// 0.01秒ごとに現在の値を取得し、50なら0にします
			swapped := atomic.CompareAndSwapInt64(&i, 50, 0)
			if swapped {
				// 値の書き換えに成功したときのみ標準出力にログを表示
				fmt.Printf("CAS now: reset zero\n")
			}

			// 0.01秒ごとに繰り返すための select
			select {
			case <-ctx.Done():
				return
			case <-timer:
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		// SIGUSR1 を受け取るチャネルを生成
		sig := make(chan os.Signal, 1)
		// SIGUSR1 がきたら sig チャネルに書き込む
		signal.Notify(sig, syscall.SIGUSR1)

		for {
			// context.Context が終了した場合はループを脱出し
			// sig チャネルに受信すれば処理を実行
			select {
			case <-ctx.Done():
				return
			case <-sig:
			}

			// このサンプルコード実行中に任意のタイミングで SIGUSR1 をプロセスに送信すると
			// 現在の値を0にした上でその時点での値を標準出力に表示します
			old := atomic.SwapInt64(&i, 0)
			fmt.Printf("SIGUSR1: reset zero: old: %d\n", old)
		}
	}()

	// すべての処理の終了を待つ
	wg.Wait()
}