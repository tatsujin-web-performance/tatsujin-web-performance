# 5章 リバースプロキシの利用

## 5章3節 nginxとは

### リスト1 /etc/nginx/nginx.confの設定

```nginx
include /etc/nginx/conf.d/*.conf;
include /etc/nginx/sites-enabled/*;
```

### リスト2 /etc/nginx/sites-available/isucon.confの設定

```nginx
server {
  listen 80;

  client_max_body_size 10m;
  root /home/isucon/private_isu/webapp/public/;

  location / {
    proxy_set_header Host $host;
    proxy_pass http://localhost:8080;
  }
}
```

https://github.com/catatsuy/private-isu/blob/master/provisioning/image/files/etc/nginx/sites-available/isucon.conf

### リスト3 静的ファイルの配信をnginx経由で行う

```nginx
server {
  listen 80;

  # 省略

  location /css/ {
    root /home/isucon/private_isu/webapp/public/;
  }

  location /js/ {
    root /home/isucon/private_isu/webapp/public/;
  }

  location / {
    proxy_set_header Host $host;
    proxy_pass http://localhost:8080;
  }
}
```

### リスト4 expiresの設定

```nginx
  location /css/ {
    root /home/isucon/private_isu/webapp/public/;
    expires 1d;
  }
```

## 5章5節 nginxによる転送時のデータ圧縮

### リスト5 gzip圧縮を利用する場合の設定

```nginx
gzip on;
gzip_types text/css text/javascript application/javascript application/x-javascript application/json;
gzip_min_length 1k;
```

## 5章7節 nginxによるEarly Hintsの活用

### コード例 Goで103 Early Hintsを返す

```go
package main

import (
  "net/http"
)

func earlyHintsHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Link", "</css/ea.css>; rel=preload; as=style")
  w.Header().Add("Link", "</js/ea.js>; rel=preload; as=script")

  // 103（情報レスポンス）を送る
  w.WriteHeader(http.StatusEarlyHints)

  // 最終レスポンス（200）を返す
  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  w.Header().Set("Cache-Control", "private")

  w.WriteHeader(http.StatusOK)

  w.Write([]byte("<!doctype html><html><head></head><body>Hello</body></html>"))
}

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", earlyHintsHandler)
  http.ListenAndServe("127.0.0.1:8000", mux)
}
```

### 設定例 nginxでEarly Hintsを有効にする

```nginx
upstream easerver {
  server 127.0.0.1:8000;
}

server {
  ...（省略）

  location = / {
    # クライアント側：HTTP/2またはHTTP/3のリクエストにだけ103を転送する
    early_hints $http2$http3;

    proxy_pass http://easerver;
  }
}
```

## 5章8節 nginxとアップストリームサーバーのコネクション管理

### リスト6 nginx 1.29.7より前でアップストリームサーバーとのコネクションを保持する設定

```nginx
location / {
  proxy_http_version 1.1;
  proxy_set_header Connection "";
  proxy_pass http://app;
}
```

### リスト7 keepalive関連のパラメータを明示的に指定する

```nginx
upstream app {
  server localhost:8080;

  keepalive 32;
  keepalive_requests 10000;
}
```

## コラム：更なるnginx高速化

### リスト8 sendfileとtcp_nopushを両方とも有効にする設定

```nginx
sendfile on;
tcp_nopush on;
```
