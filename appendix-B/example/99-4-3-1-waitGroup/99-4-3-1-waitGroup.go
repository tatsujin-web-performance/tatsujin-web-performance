package main

import (
  "fmt"
  "sync"
  "time"
)

func main() {
  // sync.WaitGroup の生成
  wg := &sync.WaitGroup{}

  // このコード例では待ち合わせする処理が2つなのが確定しているので、wg.Add の引数に2を渡しています
  // wg.Add した数以上 wg.Done を呼び出すと panic が発生するので気をつけてください
  wg.Add(2)

  // ループ中で goroutine を生成する場合などは各生成時に wg.Add(1) を呼び出すとよいでしょう
  // wg.Add(1)
  go func() {
    // よくある間違いとして goroutine の中で wg.Add(1) してしまうケースがありますが
    // その場合 goroutine が起動する前に wg.Wait() に到達してまう場合があるので
    // gorountine の中で wg.Add(1) しないように注意

    // 処理が終了し関数を抜ける際に確実に wg.Done されるように先頭で defer を使って wg.Done を呼び出しています
    defer wg.Done()

    // 5秒間毎秒標準出力へ表示
    for i := 0; i < 5; i++ {
      fmt.Printf("wg 1: %d / 5\n", i+1)
      time.Sleep(1 * time.Second)
    }
  }()

  // wg.Add(1)
  go func() {
    // 処理が終了し関数を抜ける際に確実に wg.Done されるように先頭で defer を使って wg.Done を呼び出しています
    defer wg.Done()

    // 5秒間毎秒標準出力へ表示
    for i := 0; i < 5; i++ {
      fmt.Printf("wg 2: %d / 5\n", i+1)
      time.Sleep(1 * time.Second)
    }
  }()

  // ここで2つの goroutine が終了して wg.Done が呼び出されるのを待っています
  // wg.Wait の返り値はなく、チャネルで終了通知の受信などは出来ないため気をつけてください
  wg.Wait()

  fmt.Println("wg: done")
}
