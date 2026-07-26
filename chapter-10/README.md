# 10章 分散トレーシング

10章「分散トレーシング」のサンプルコードです。

フロントエンドとバックエンドの2つのサービスで構成されるメモ保存アプリケーションに、OpenTelemetryを使用して分散トレーシングの計装を行います。

```
User → Frontend(:8080) → Backend(:8081) → SQLite
```

## Goアプリケーションへの計装

- [`example`](example) 手動計装を行ったサンプルアプリケーション
  - [`example/frontend`](example/frontend) フロントエンドサービス（認証処理・バックエンド呼び出しのスパン作成、トレースコンテキストの注入）
  - [`example/backend`](example/backend) バックエンドサービス（トレースコンテキストの抽出、HTTPハンドラとDB保存処理のスパン作成）
- [`example-auto`](example-auto) 自動計装ライブラリを使用したサンプルアプリケーション
  - [`example-auto/frontend`](example-auto/frontend) フロントエンドサービス（`otelhttp.NewHandler` / `otelhttp.NewTransport` による自動計装、認証処理のみ手動計装）
  - [`example-auto/backend`](example-auto/backend) バックエンドサービス（`otelhttp.NewHandler` / `otelsql` による自動計装）

### 動作確認方法

各ディレクトリでDocker Composeを使って、frontend、backend、Jaegerの3つのサービスを起動できます。

```console
$ cd example
$ docker compose up --build
```

アプリケーションが起動したら、別のターミナルからfrontendサービスにリクエストを送信します。

```console
$ curl -X POST http://localhost:8080/memos \
    -H "Authorization: Bearer token-user123" \
    -H "Content-Type: application/json" \
    -d '{"content":"This is a sample memo."}'
{"status":"created"}
```

### Jaeger UIでのトレース確認

リクエストを送信した後、ブラウザで http://localhost:16686 にアクセスするとJaeger UIが表示されます。「Service」ドロップダウンで「frontend」または「backend」を選択し、「Find Traces」ボタンをクリックすると、記録されたトレースを確認できます。

### トレースを標準出力に出力する

サンプルアプリケーションは、環境変数 `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` が設定されている場合はOTLPエクスポーターでJaegerにトレースを送信し、未設定の場合は標準出力にJSON形式でトレースを出力します。

[`compose.yaml`](example/compose.yaml) の `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` をコメントアウトして起動すると、`docker compose up` を実行したターミナルにトレースが出力されます。

### Docker Composeを使わずに実行する

Goがインストールされていれば、各サービスを直接起動することもできます（backendは [`mattn/go-sqlite3`](https://github.com/mattn/go-sqlite3) を使用しているため、CGOが有効である必要があります）。

```console
$ cd example/backend && go run . &
$ cd example/frontend && go run . &
```

環境変数 `BACKEND_URL` でfrontendからbackendへの接続先を、`OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` でトレースの送信先を変更できます。

### 補足: semconvパッケージのバージョンについて

本サンプルコードでは `go.opentelemetry.io/otel/semconv/v1.41.0` をimportしています。`resource.Default()` とマージするリソースのスキーマURLは、使用しているOpenTelemetry SDKが内部で使用するsemconvのバージョンと一致している必要があり、一致しない場合は `resource.Merge()` が `conflicting Schema URL` エラーを返します。SDKのバージョンを変更した場合は、semconvのimportバージョンも合わせて変更してください。
