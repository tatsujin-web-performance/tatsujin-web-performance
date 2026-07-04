# 4章 シナリオを持った負荷試験

4章「シナリオを持った負荷試験」のサンプルコードです。

## 4-2 k6による単純な負荷試験

- リスト1 単一URLにリクエストを送信するシナリオ [`ab.js`](ab.js)

## 4-3 k6でシナリオを記述する

- リスト2 対象 URL を生成する関数を定義した [`config.js`](config.js)
- リスト3 Web サービスの初期化処理を行うシナリオ [`initialize.js`](initialize.js)
- リスト4 ユーザーがログインしてコメントを投稿するシナリオ [`comment.js`](comment.js)
- リスト6  ログインしてフォームから画像をアップロードするシナリオ [`postimage.js`](postimage.js)
- リスト7 アカウント情報を定義する JSON ファイル [`accounts.json`](accounts.json)
- リスト8 [`accounts.json`](accounts.json) を `SharedArray` として読み込むモジュール [`accounts.js`](accounts.js)

(注) 「リスト9 [`accounts.js`](accounts.js) を `import` して `getAccount()` 関数を利用するシナリオ」にある変更を [`comment.js`](comment.js) と [`postimage.js`](postimage.js) に対して適用する場合は、このディレクトリで以下のコマンドを実行してpatchを適用してください。

```console
$ patch -p2 < comment.js.patch
$ patch -p2 < postimage.js.patch
```

## 4-4 複数のシナリオを組み合わせた統合シナリオを実行する

- リスト10 複数のシナリオ関数を組み合わせて実行する [`integrated.js`](integrated.js)

(注) [`integrated.js`](integrated.js) は [`comments.js`](comments.js) と [`postimage.js`](postimage.js) に上記のpatchが適用済みであることを前提にしています。
