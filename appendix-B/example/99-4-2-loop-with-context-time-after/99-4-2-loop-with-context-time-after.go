package main

import (
  "context"
  "fmt"
  "time"
)

func main() {
  // 5.5秒でタイムアウトする context.Context を作る
  ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second + 500 * time.Millisecond)
  defer cancel()

  i := 0

L: // ループ脱出用のラベル
  for {
    // ループごとに出力
    // time.After なら loop 5 までしか出力されないが、time.Sleep の場合1つ余計に loop 6 まで出力される
    fmt.Printf("loop %d\n", i)
    i++

    select {
      // ctx が終了していれば L ラベルまで脱出して for ループを抜ける
      // 単に break と書くと select の break になってしまい無限ループが継続するので注意
      case <-ctx.Done():
        break L
      // ctx が終了していなければ1秒待つが、チャネルの受信にしているので先に ctx が終了すればそちらが実行される
      case <-time.After(1 * time.Second):

      // time.Sleep で待つ例
      // default:
      //   time.Sleep(1 * time.Second)
    }
  }
}
