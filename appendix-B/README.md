# 付録B ISUCON ベンチマーカー実装

付録 B 「ISUCON ベンチマーカー実装」のサンプルコードです。

## ベンチマーカーに頻出する実装パターン

### `context.Context` の生成

- [99-4-1-1-context.go](./example/99-4-1-1-context/99-4-1-1-context.go)

#### `context.TODO` の使い所

- [99-4-1-1-c-context-todo.go](./example/99-4-1-1-c-context-todo/99-4-1-1-c-context-todo.go)


#### `context.WithCancel` による `context.Context` の中断

- [99-4-1-3-with-cancel.go](./example/99-4-1-3-with-cancel/99-4-1-3-with-cancel.go)

#### `context.WithTimeout(ctx, d)` と `context.WithDeadline(ctx, t)` による時限式の `context.Context` の中断

- [99-4-1-4-with-timeout.go](./example/99-4-1-4-with-timeout/99-4-1-4-with-timeout.go)

### `time` と `context` によるループのパターン

- [99-4-2-loop-with-context.go](./example/99-4-2-loop-with-context/99-4-2-loop-with-context.go)
- [99-4-2-loop-with-context-time-after.go](./example/99-4-2-loop-with-context-time-after/99-4-2-loop-with-context-time-after.go)
- [99-4-2-loop-with-context-time-after-long-time.go](./example/99-4-2-loop-with-context-time-after-long-time/99-4-2-loop-with-context-time-after-long-time.go)

### `sync` パッケージの利用

#### `sync.WaitGroup` による待ち合わせ

- [99-4-3-1-waitGroup.go](./example/99-4-3-1-waitGroup/99-4-3-1-waitGroup.go)

#### `sync.Mutex` と `sync.RWMutex` による読み書きのロック

- [99-4-3-2-mutex.go](./example/99-4-3-2-mutex/99-4-3-2-mutex.go)

#### `sync.WaitGroup` や `sync.Mutex` を値渡しすることで発生するデッドロックや panic

- [99-4-3-3-panic-with-mutex.go](./example/99-4-3-3-panic-with-mutex/99-4-3-3-panic-with-mutex.go)

## private-isu を対象としたベンチマーカーの実装

こちらのソースコードは別途リポジトリに配置されています。

https://github.com/rosylilly/private-isu-benchmarker/

##### (コラム) `fmt.Stringer` と `fmt.GoStringer` を実装する

- [99-5-2-c-stringer.go](./example/99-5-2-c-stringer/99-5-2-c-stringer.go)
