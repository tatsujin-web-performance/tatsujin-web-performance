package main

import (
  "context"
  "fmt"
)

func main() {
  // 一般的に関数内で生成した context.CancelFunc は defer などで確実に context.Context が終了するように書く
  ctxParent, cancelParent := context.WithCancel(context.Background())
  defer cancelParent()

  ctxChild, cancelChild := context.WithCancel(ctxParent)
  defer cancelChild()

  // 親 context.Context の中断は子 context.Context にも伝播する
  // context.CancelFunc は何度呼び出してもよい(2度目以降の呼び出しは何もしない)
  cancelParent()
  // 親 context.Context に中断が伝播しないことを確認する場合は直上1行をコメントアウトして、直下1行のコメントアウトを解除する
  // cancelChild()

  // context.Canceled が返る。子 context.Context の context.CancelFunc を実行しただけの場合は nil
  fmt.Printf("parent.Err is %v\n", ctxParent.Err())
  // => parent.Err is context canceled

  // 親 context.Context の中断が伝播し、子 context.Context でも context.Canceled になる
  fmt.Printf("child.Err is %v\n", ctxChild.Err())
  // => child.Err is context canceled
}
