package main

import (
  "context"
  "fmt"
  "time"
)

func main() {
  // 10秒でタイムアウトする context.Context を作る
  ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
  defer cancel()

  go LoopWithBefore(ctx)
  go LoopWithAfter(ctx)

  <-ctx.Done()
}

// ループの最初で time.After を生成して待ち合わせるパターン
// HeavyProcess に1.5秒かかるが、1ループの時間は3秒に収まっている
func LoopWithBefore(ctx context.Context) {
  // ループ前の時間を取得
  beforeLoop := time.Now()
  for {
    // 1ループの持ち時間を先頭で設定
    loopTimer := time.After(3 * time.Second)

    // 1.5秒かかる処理
    HeavyProcess(ctx, "BEFORE")

    select {
      case <-ctx.Done():
        return
      // 先頭で生成した time.After を使って待ち合わせする
      case <-loopTimer:
        // 1ループにかかった時間を標準出力に表示して、 beforeLoop に現在時刻を設定
        fmt.Printf("[BEFORE] loop duration: %.2fs\n", time.Now().Sub(beforeLoop).Seconds())
        beforeLoop = time.Now()
    }
  }
}

// ループの最後で time.After を生成して待ち合わせるパターン
// HeavyProcess で1.5秒かかり、その上で3秒待つため、1ループの時間は合計4.5秒になっている
func LoopWithAfter(ctx context.Context) {
  beforeLoop := time.Now()
  for {
    // 1.5秒かかる処理
    HeavyProcess(ctx, "AFTER")

    select {
      case <-ctx.Done():
        return
      // この場で生成した time.After を使って待ち合わせする
      case <-time.After(3 * time.Second):
        // 1ループにかかった時間を標準出力に表示して、 beforeLoop に現在時刻を設定
        fmt.Printf("[AFTER] loop duration: %.2fs\n", time.Now().Sub(beforeLoop).Seconds())
        beforeLoop = time.Now()
    }
  }
}

// どちらのループから呼び出されたのかを表示しつつ、1.5秒待つ
func HeavyProcess(ctx context.Context, pattern string) {
  fmt.Printf("[%s] Heavy Process\n", pattern)
  time.Sleep(1 * time.Second + 500 * time.Millisecond)
}
