# 8章 押さえておきたい高速化手法

## 8章1節 外部コマンド実行ではなくライブラリを利用する

### リスト1 opensslコマンドを実行するGoの初期実装

```go
func digest(src string) string {
  out, err := exec.Command("/bin/bash", "-c", `printf "%s" `+escapeshellarg(src)+` | openssl dgst -sha512 | sed 's/^.*= //'`).Output()
  （省略）
```

https://github.com/catatsuy/private-isu/blob/0c9a8f258c759d5133c6200c6453f82703663614/webapp/golang/app.go#L122-L131

### リスト2 opensslコマンドを実行するRubyの初期実装

```ruby
def digest(src)
  `printf "%s" #{Shellwords.shellescape(src)} | openssl dgst -sha512 | sed 's/^.*= //'`.strip
end
```

https://github.com/catatsuy/private-isu/blob/0c9a8f258c759d5133c6200c6453f82703663614/webapp/ruby/app.rb#L78-L81

### リスト3 Goの実装

```go
import (
  "fmt"
  "crypto/sha512"
（省略）
)

func digest(src string) string {
  return fmt.Sprintf("%x", sha512.Sum512([]byte(src)))
}
```

### リスト4 Rubyの実装

```ruby
require 'openssl'
（省略）

def digest(src)
  return OpenSSL::Digest::SHA512.hexdigest(src)
end
```

## コラム：実装する言語によって高速になるのか

### リスト5 strings.NewReplacerの利用

```go
r := strings.NewReplacer("<", "&lt;", ">", "&gt;")
fmt.Println(r.Replace("This is <b>HTML</b>!")) // This is &lt;b&gt;HTML&lt;/b&gt;!
```

## 8章2節 開発用の設定で冗長なログを出力しない

### リスト6 デバッグモードの無効、ログレベルを変更

```diff
 func main() {
  e := echo.New()
- e.Debug = true
- e.Logger.SetLevel(log.DEBUG)
+ e.Debug = false
+ e.Logger.SetLevel(log.ERROR)
```

https://github.com/isucon/isucon11-qualify/blob/1011682c2d5afcc563f4ebf0e4c88a5124f63614/webapp/go/main.go#L211-L212

## 8章3節 HTTPクライアントの使い方

### リスト7 res.Body.Close()を実行して、レスポンスのBodyを読み切る

```go
res, err := http.DefaultClient.Do(req)
if err != nil {
  log.Fatal(err)
}
defer res.Body.Close()

_, err = io.ReadAll(res.Body)
if err != nil {
  log.Fatal(err)
}
```

### リスト8 Timeoutを指定する

``` go
hClient := http.Client{
  Timeout: 5 * time.Second,
}
```

### リスト9 http.Transportで確認した方が良い設定

``` go
hClient := http.Client{
  Timeout:   5 * time.Second,
  Transport: &http.Transport{
    MaxIdleConns:        500,
    MaxIdleConnsPerHost: 200,
    IdleConnTimeout:     120 * time.Second,
  },
}
```

## 8章4節 静的ファイル配信をリバースプロキシから直接配信する

### リスト10 /home/isucon/private_isu/webapp/public/image/ディレクトリ上に画像ファイルを配置する

```nginx
server {
  # 省略
  location /image/ {
    root /home/isucon/private_isu/webapp/public/;
    try_files $uri @app;
  }

  location @app {
    proxy_pass http://localhost:8080;
  }
```

## 8章5節 HTTPヘッダーを活用してクライアント側にキャッシュさせる

### リスト11 Cache-Controlヘッダーをレスポンスに含む設定

```nginx
server {
  # 省略
  location /image/ {
    root /home/isucon/private_isu/webapp/public/;
    expires 1d;
  }
```
