# 付録A private-isu攻略実践

付録A「private-isu攻略実践」のサンプルコードです。

このディレクトリの private-isu/webapp 以下に、[catatsuy/private-isu](https://github.com/catatsuy/private-isu) を付録の記述内容に従って改変したコードの例が保存されています。

### unicorn workerプロセスを4にする(約15,000点)

- [リスト1 編集したunicorn_config.rb](private-isu/webapp/ruby/unicorn_config.rb)
- [commit](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/83a57020b2e205c8a7d6163ee3c58fed361f6605)

### 静的ファイルをnginxで配信する(17,000点)

- [リスト2 /etc/nginx/sites-available/isucon.conf](private-isu/webapp/etc/nginx/conf.d/default.conf)
- [commit](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/9f34d0bd90146a050cc4c7e13d5c89743c7f77be)

### アップロード画像を静的ファイル化する(約22,000点)

- [リスト 3 アップロード画像を順次静的ファイルに移行するアプリケーションの改修](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/907c70b3f7e722068fd3462445d8cee8efb27a76)
- [リスト 4 /image/ 以下の静的ファイルがあれば配信、なければアプリケーションサーバーにリバースプロキシするnginxの設定](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/a49eb2dfd307ffe7a85eb9cfbcae3912e8427f0d)


### postsとusersをJOINして必要な行数だけ取得する(約90,000点)

- [リスト9, 10の差分](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/7f73caf78e714e982b9f24478e0919b8e50af2b6)

### プリペアードステートメントの改善(約110,000点)

- [リスト12 prepare.execteをxqueryに変更する例](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/57592ee2681fc3551ab810932b1706fc775aac43)

### postsからのN+1クエリ結果をキャッシュ(約180,000点)

- [リスト 13 make_postsの中でキャッシュを扱うコードの例](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/8de837130c50186ce5cd08b560552cd97a1e9b34)

### 適切なインデックスが使えないクエリを解決(約200,000点)

- [STRAIGHT_JOINとFORCE INDEXを追加した例](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/8bf1580d4542dd7e4c798dcc6c6aead6cd0bd339)


### 外部コマンド呼び出しをやめる(約240,000点)

- [opensslコマンドの呼び出しをやめてopenssl gemを使う](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/8f3b4a8839a3583a8abd2bde0cea2e2bfa2f8c20)

### MySQLの設定調整(約255,000点)

- [MySQLの設定調整](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/d9337bbf0031318fa25360befdca86459920e780)

### memcachedへのN+1解消(約300,000点)

- [memcachedへのN+1をget_multiで解消](https://github.com/tatsujin-web-performance/tatsujin-web-performance/commit/7885e234db464078e973921c03fcb8968a776d52)
