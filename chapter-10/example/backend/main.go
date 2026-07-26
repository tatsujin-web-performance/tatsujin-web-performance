package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"
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
	// HTTPヘッダーからトレースコンテキストを抽出
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	tracer := otel.Tracer("backend")
	ctx, span := tracer.Start(ctx, "handleInternalCreateMemo",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	// リクエストの属性を記録
	span.SetAttributes(
		semconv.HTTPRequestMethodKey.String(r.Method),
		semconv.URLPath(r.URL.Path),
	)

	// リクエストのbodyを構造体にデコード
	var req InternalMemoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decode request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// リクエストのユーザーIDを属性として記録
	span.SetAttributes(attribute.String("user_id", req.UserID))

	// ctxを渡すことで親子関係が構築される
	if err := saveMemo(ctx, req.UserID, req.Content); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save memo")
		http.Error(w, "failed to save memo", http.StatusInternalServerError)
		return
	}

	// 正常終了時はOKステータスを設定
	span.SetStatus(codes.Ok, "")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, `{"status":"created"}`)
}

func saveMemo(ctx context.Context, userID, content string) error {
	// 親のコンテキストから子スパンを作成
	tracer := otel.Tracer("backend")
	ctx, span := tracer.Start(ctx, "saveMemo",
		trace.WithSpanKind(trace.SpanKindClient), // DBアクセスはクライアントとして扱う
	)
	defer span.End()

	query := "INSERT INTO memos (user_id, content) VALUES (?, ?)"

	// DB操作の属性を記録
	span.SetAttributes(
		semconv.DBSystemNameSQLite,
		semconv.DBQueryText(query),
	)

	if _, err := db.ExecContext(ctx, query, userID, content); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to insert memo")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func initDB(path string) error {
	var err error
	db, err = sql.Open("sqlite3", path)
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

	slog.Info("Backend server starting on :8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
