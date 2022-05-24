package main

import (
  "fmt"
  "sync"
  "time"
)

func main() {
  // sync.Mutex を値として生成
  var mu sync.Mutex

  // mu をロック
  mu.Lock()

  // 1秒後に mu をアンロック
  go func() {
    <-time.After(1 * time.Second)
    mu.Unlock()
  }()

  // 値渡しだとコピーされてしまい1秒後の Unlock が関数呼び出し先の mu に伝わらないため、
  // 関数内の Lock がいつまでも解除されず deadlock として Go ランタイムが強制終了
  LockWithValue(mu)

  // 参照渡しの場合は1秒後の Unlock が期待通り関数内の mu に伝わるため、
  // deadlock は発生せず、1秒後に mutex unlocked が標準出力に表示されてプログラムが正常終了
  // LockWithReference(&mu)

  fmt.Println("mutex unlocked")
}

// 値渡しで sync.Mutex を受け取る関数
func LockWithValue(mu sync.Mutex) {
  mu.Lock()
  mu.Unlock()
}

// 参照渡しで sync.Mutex を受け取る関数
func LockWithReference(mu *sync.Mutex) {
  mu.Lock()
  mu.Unlock()
}
