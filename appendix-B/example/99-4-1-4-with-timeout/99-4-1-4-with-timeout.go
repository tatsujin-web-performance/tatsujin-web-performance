package main

import (
  "context"
  "fmt"
  "time"
)

func main() {
  ctxMain := context.Background()

  go func() {
    // 5秒後にタイムアウトする context.Context を生成
    ctxTimeout, cancelTimeout := context.WithTimeout(ctxMain, 5 * time.Second)
    // context.CancelFunc を呼び出して解放するのを忘れないように
    defer cancelTimeout()

    // context.Context の終了を待つ
    <-ctxTimeout.Done()

    // ちょうど5秒後に出力
    fmt.Println("timeout!")
  }()

  go func() {
    // 3秒後にタイムアウトする context.Context を生成
    ctxDeadline, cancelDeadline := context.WithDeadline(
      ctxMain,
      // 現在時刻に3秒足す
      time.Now().Add(3 * time.Second),
    )
    // context.CancelFunc を呼び出して解放するのを忘れないように
    defer cancelDeadline()

    // context.Context の終了を待つ
    <-ctxDeadline.Done()

    // ちょうど3秒後に出力
    fmt.Println("deadline!")
  }()

  // 10秒間毎秒ごとに n sec...と標準出力に表示するコード
  for i := 0; i < 10; i++ {
    fmt.Printf("%d sec...\n", i)
    time.Sleep(1 * time.Second)
  }
}
