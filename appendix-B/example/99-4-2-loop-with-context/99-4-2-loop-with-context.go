package main

import (
  "context"
  "fmt"
  "time"
)

func main() {
  // 5秒でタイムアウトする context.Context を作る
  ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
  defer cancel()

L: // ループ脱出用のラベル
  for {
    // ループごとに出力
    fmt.Println("loop")

    select {
      // ctx が終了していれば L ラベルまで脱出して for ループを抜ける
      // 単に break と書くと select の break になってしまい無限ループが継続するので注意
      case <-ctx.Done():
        break L
      // ctx が終了していなければ1秒待つ
      default:
        time.Sleep(1 * time.Second)
    }
  }
}
