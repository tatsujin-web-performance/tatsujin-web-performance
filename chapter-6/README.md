# 6章 リバースプロキシの利用

## 6章3節 nginxについて

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

## 6章5節 nginxによる転送時のデータ圧縮

### リスト5 gzip圧縮を利用する場合の設定

```nginx
gzip on;
gzip_types text/css text/javascript application/javascript application/x-javascript application/json;
gzip_min_length 1k;
```

## 6章7節 nginxとアップストリームサーバーのコネクション管理

### リスト6 アップストリームサーバーとのコネクションを保持する設定

```nginx
location / {
  proxy_http_version 1.1;
  proxy_set_header Connection "";
  proxy_pass http://app;
}
```

### リスト7 keepaliveとkeepalive_requestsを利用する

```nginx
upstream app {
  server localhost:8080;

  keepalive 32;
  keepalive_requests 10000;
}
```

## コラム：更なるnginx高速化

```nginx
sendfile on;
tcp_nopush on;
```
