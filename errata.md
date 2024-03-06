# 正誤表

## 1版 1刷→2刷での修正箇所

- p.19 表3 負荷試験の概要 シナリオ
  - 誤)例： ログイン→カレンダー表示→予約枠選択→確認画面→予約実行という行動をとるユーして1回やりなおすユーザーが5%
  - 正)例： ログイン→カレンダー表示→予約枠選択→確認画面→予約実行という行動をとるユーザーが全体の95%、その後にキャンセルして1回やりなおすユーザーが5%

- p.24 注1
  - 誤）http://developer.cybozu.co.jp/archives/kazuho/2010/01/cronlog-52f2.html
  - 正）https://b.hatena.ne.jp/entry/developer.cybozu.co.jp/archives/kazuho/2010/01/cronlog-52f2.html

- p.29 本文13行目
  - 誤）humanreable
  - 正）human-readable

- p.48 本文9行目
  - 誤）Monitoing
  - 正）Monitoring

- p.52 本文11行目
  - 誤）C5.large
  - 正）c5.large

- p.74 本文2行目
  - 誤）C5.large
  - 正）c5.large

- p.77 本文7行目
  - 誤）Request per second
  - 正）Requests per second

- p.84 本文7行目
  - 誤）Request per second
  - 正）Requests per second

- p.92 コラム4行目
  - 誤）C5.large
  - 正）c5.large

- p.118 本文6行目
  - 誤）Web2.0
  - 正）Web 2.0

- p.170 本文6行目
  - 誤）lua
  - 正）Lua

- p.231 本文21行目
  - 誤）Receive-side Scaling
  - 正）Receive Side Scaling

- p.266 本文13行目
  - 誤）C6i.large
  - 正）c6i.large

- p.266 本文18行目
  - 誤）C6i.large
  - 正）c6i.large

- p.340 索引
  - 誤）kTLS.......180
  - 正）kTLS.......180,242

## 書籍版未修正箇所

- p.19 表3
  - 誤) 予約実行という行動をとるユーして1回やり直すユーザーが5%
  - 正) 予約実行という行動をとるユーザーが全体の95%、その後にキャンセルして1回やりなおすユーザーが5%

- p.84 本文9行目
  - 誤) 50回
  - 正) 20回

- p.100 本文1行目
  - 誤) /ininitlize
  - 正) /initialize

- p.141 リスト8
  - 誤) `WHERE comments`
  - 正) `WHERE comment`

- p.147 リスト12

```diff
 if err == nil {
     // ユーザー情報があればJSONをデコードして返す
     err := json.Unmarshal(it.Value, &user)
-    if err != nil {
+    if err == nil {
         return user
     }
  }
```

- p.149 リスト13

```diff
     users := map[int]User{}
     // キャッシュから複数のキャッシュを取得 ❸
     items, err := mc.GetMulti(keys)
-    if err == nil {
+    if err != nil {
         return users
     }
     for _, it := range items {
```
