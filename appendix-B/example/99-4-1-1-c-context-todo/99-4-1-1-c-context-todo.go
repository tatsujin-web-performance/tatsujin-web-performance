package main

import (
  "context"
)

func main() {
  // context.Context をサポートしていない関数の呼び出し
  // 本当はこの main 関数内で context.Context を生成して渡したい
  ContextNotSupportedFunc()
}

// 引数に context.Context を取らない関数
// 後ほど context.Context を引数に受け取るように変更される
func ContextNotSupportedFunc() {
  // 渡すべき context.Context がないので一旦暫定的に context.TODO で生成して渡す
  RequiredContextFunc(context.TODO())
}

// 引数に context.Context を必要とする関数
func RequiredContextFunc(ctx context.Context) {
}
