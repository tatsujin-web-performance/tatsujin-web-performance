package main

import (
  "fmt"
  "sync"
)

func main() {
  // int のスライスを生成
  userIDs := []int{}
  // sync.Mutex を生成
  userIDsLock := &sync.Mutex{}

  // 処理の待ち合わせに利用する sync.WaitGroup の生成
  wg := &sync.WaitGroup{}

  for i := 0; i < 20; i++ {
    wg.Add(1)
    go func(id int) {
      defer wg.Done()

      // userIDs への書き込みの競合を防ぐためにロック
      // 別の goroutine ですでにロックされている場合はそのロックが解除するまでここで処理がブロック
      userIDsLock.Lock()
      // データをスライスへ追加
      userIDs = append(userIDs, id)
      // ロックの解除
      userIDsLock.Unlock()
    }(i)
  }

  // すべての追加処理を待つ
  wg.Wait()

  // 追加されたすべての値を表示
  // goroutine は開始した順に実行される訳ではないので、実行するたびに追加順が違っている
  fmt.Printf("userIDs: %v\n", userIDs)
}
