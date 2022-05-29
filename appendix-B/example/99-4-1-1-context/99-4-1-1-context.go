package main

import (
  "context"
)

func main() {
  // main 関数の先頭で新しい context.Context を生成
  ctx := context.Background()

  // context を利用する処理
  ExampleContextFunc(ctx)
}

func ExampleContextFunc(ctx context.Context) {
  // この関数内では新しい context.Context は生成せず、
  // 別関数が context.Context を必要とするなら受け取った ctx を渡す
}
