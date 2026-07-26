package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/XSAM/otelsql"
	_ "github.com/mattn/go-sqlite3"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

var db *sql.DB

func initTracer() (func(), error) {
	var exporter sdktrace.SpanExporter
	var err error

	// 環境変数でOTLPエンドポイントが指定されている場合はOTLPエクスポーターを使用
	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"); endpoint != "" {
		exporter, err = otlptracehttp.New(context.Background())
		if err != nil {
			return nil, err
		}
	} else {
		// 環境変数が未設定の場合は標準出力に出力
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
	}

	// リソース情報を設定（サービス名など）
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("backend"),
		),
	)
	if err != nil {
		return nil, err
	}

	// TracerProviderを作成
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp) // グローバルに登録

	// コンテキスト伝播の設定
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// シャットダウン関数を返す
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer provider", "error", err)
		}
	}, nil
}

// リクエストボディの構造体
type InternalMemoRequest struct {
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

func handleInternalCreateMemo(w http.ResponseWriter, r *http.Request) {
	// r.Context()にはotelhttp.NewHandlerで作成されたスパンのコンテキストが含まれる
	// 手動でスパンを作成・終了する必要がない

	// リクエストのbodyを構造体にデコード
	var req InternalMemoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// コンテキストを渡せば、子スパンとして関連付けられる
	if err := saveMemo(r.Context(), req.UserID, req.Content); err != nil {
		http.Error(w, "failed to save memo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, `{"status":"created"}`)
}

func saveMemo(ctx context.Context, userID, content string) error {
	// otelsqlにより、ExecContextの呼び出しが自動的に計装される
	_, err := db.ExecContext(ctx, "INSERT INTO memos (user_id, content) VALUES (?, ?)", userID, content)
	return err
}

func initDB(path string) error {
	var err error
	// otelsqlでデータベース接続をラップ
	db, err = otelsql.Open("sqlite3", path,
		otelsql.WithAttributes(semconv.DBSystemNameSQLite),
	)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS memos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

func main() {
	shutdown, err := initTracer()
	if err != nil {
		slog.Error("failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer shutdown()

	if err := initDB("./memos.db"); err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /internal/memos", handleInternalCreateMemo)

	// otelhttp.NewHandlerでHTTPサーバーをラップ
	handler := otelhttp.NewHandler(mux, "backend")

	slog.Info("Backend server starting on :8081")
	if err := http.ListenAndServe(":8081", handler); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
