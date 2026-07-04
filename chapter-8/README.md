# 9章 OSの基礎知識とチューニング

## 9-8 Linuxカーネルパラメータ

### リスト1 UNIX domain socketを用いた例

```nginx
server {

  ## 80番ポートで接続を待機する際の設定( # を付けてコメントアウト済)
  # listen 80;

  ## /var/run/nginx.sock で接続を待機する際の設定
  listen unix:/var/run/nginx.sock;

<以下略>
```

### リスト2 unicorn_config.rb にある設定

```ruby
worker_processes 1
preload_app true
listen "0.0.0.0:8080"
```

### リスト3 /tmp/webapp.sock に変更

```ruby
worker_processes 1
preload_app true
listen "/tmp/webapp.sock"
```

### リスト4 Go実装を書き換える

```go
## "/tmp/webapp.sock" で listen(2) する
listener, err := net.Listen("unix", "/tmp/webapp.sock")
if err != nil {
        log.Fatalf("Failed to listen on /tmp/webapp.sock: %s.", err)
}
defer func() {
        err := listener.Close()
        if err != nil {
                log.Fatalf("Failed to close listener: %s.", err)
        }
}()

## systemdなどから送信されるシグナルを受け取る
c := make(chan os.Signal, 2)
signal.Notify(c, os.Interrupt, syscall.SIGTERM)
go func() {
        <-c
        err := listener.Close()
        if err != nil {
                log.Fatalf("Failed to close listener: %s.", err)
        }
}()

log.Fatal(http.Serve(listener, mux))
```

### リスト5 初期状態の設定

```nginx
server {
<省略>
  location / {
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_pass http://localhost:8080;
  }
}  
```

### リスト6 /tmp/webapp.sock をアップストリームサーバーとして指定した設定

```nginx
upstream webapp {
  server unix:/tmp/webapp.sock;
}

server {
<省略>
  location / {
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_pass http://webapp;
  }
}
```
