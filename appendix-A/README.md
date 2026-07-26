# 付録A private-isu攻略実践

付録A「private-isu攻略実践」のサンプルコードです。

このディレクトリの [private-isu/](private-isu/) 以下に、[catatsuy/private-isu](https://github.com/catatsuy/private-isu) のWebアプリケーション(Go実装 `webapp/golang`)を、付録の記述内容に従って改変したコードの例が保存されています。付録内の各節に対応する変更は、それぞれ1つのcommitとして積まれています。

なお、書籍ではAmazon EC2上の競技環境(`/home/isucon/private_isu/webapp` 以下)を対象としているのに対し、このリポジトリのnginx設定などはprivate-isuに付属するDocker Compose環境(`webapp/compose.yml`)向けのパス(`/public/` や `app:8080` など)になっています。適宜読み替えてください。

MySQLに対して直接実行する `ALTER TABLE` などのSQLはcommitに含まれていないため、各節に記載しています。

## 各章の技法を適用する

### commentsテーブルにインデックスを追加する(約33,000点)

スロークエリログを有効にして解析し、`comments`テーブルにインデックスを追加します。

```sql
ALTER TABLE comments ADD INDEX post_id_idx (post_id, created_at DESC);
```

### 静的ファイルをnginxで配信する(約35,000点)

- [リスト2 nginxの設定](private-isu/webapp/etc/nginx/conf.d/default.conf)
- [commit](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/5f908ec32152f5bfda9f4f581440b2b2de2af20e)

### アップロード画像を静的ファイル化する(約51,000点)

- [リスト3 アップロード画像を順次静的ファイルに移行するアプリケーションの改修](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/36ccbc75a180a473e54d645465e8dadc5655b331)
- [リスト4 /image/ 以下の静的ファイルがあれば配信、なければアプリケーションサーバーにリバースプロキシするnginxの設定](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/01428ac2c816a7d4e925958f3872df5cd3aac1b6)

### postsとusersをJOINして必要な行数だけ取得する(約100,000点)

`ORDER BY`狙いのインデックスを追加した上で、`posts`と`users`をJOINして`LIMIT`付きで取得します。

```sql
ALTER TABLE posts ADD INDEX posts_order_idx (created_at DESC);
```

- [リスト9, 10 makePostsとクエリの変更](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/1731c9ee97650f779f242a1c7304948ee2427a01)

### プリペアードステートメントを改善する(約130,000点)

- [リスト12 interpolateParams=trueでクライアントサイドプリペアードステートメントを使う](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/cfd4569fb62c0bddb1e1af478dc262e1af5ee730)

### commentsテーブルへインデックス追加する(約145,000点)

```sql
ALTER TABLE comments ADD INDEX idx_user_id (user_id);
```

### postsからのN+1クエリ結果をキャッシュ(約240,000点)

- [リスト14 makePostsの中でキャッシュを扱うコードの例](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/98407bc3bf5e1f9d47073553014bcf532ed99bfb)

### 適切なインデックスが使えないクエリを解決(約250,000点)

複合インデックスを追加した上で、オプティマイザが適切なインデックスを選択できないクエリに`FORCE INDEX`と`STRAIGHT_JOIN`のヒントを追加します。

```sql
ALTER TABLE posts ADD INDEX posts_user_idx (user_id, created_at DESC);
```

- [STRAIGHT_JOINとFORCE INDEXを追加した例](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/5376867427cf0b881f36b4b377a5a40d595de4f8)

### 外部コマンド呼び出しをやめる(約390,000点)

- [opensslコマンドの呼び出しをやめてGoの標準ライブラリでSHA-512を計算する](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/8b0e471fac8891d7f0f5762b1c27f09971ad78e5)

### MySQLの設定を調整する(約400,000点)

- [MySQLの設定](private-isu/webapp/etc/mysql/conf.d/my.cnf)
- [commit](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/97164b3faad94e855ede2fe232dbd1428bf90dbf)

### アプリケーションとミドルウェア間の接続最適化(約480,000点)

- [nginxのupstream keepaliveとMySQL/memcachedの接続数設定](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/4e6e7e3053c351d32e70298048daf5694f4abd30)

### memcachedへのN+1解消(約630,000点)

- [リスト15 memcachedへget_multiしてN+1を解消するコード例](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/31a680d061f16acce0adfc77b86e10d266ef12d7)
