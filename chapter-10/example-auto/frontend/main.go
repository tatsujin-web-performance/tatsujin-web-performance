package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

var backendURL = "http://localhost:8081"

// otelhttp.NewTransportでラップしたHTTPクライアント
var httpClient = &http.Client{
	Transport: otelhttp.NewTransport(http.DefaultTransport),
}

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
			semconv.ServiceName("frontend"),
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

// ユーザーからのリクエストボディの構造体
type MemoRequest struct {
	Content string `json:"content"`
}

// バックエンドへのリクエストボディの構造体
type InternalMemoRequest struct {
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

func handleCreateMemo(w http.ResponseWriter, r *http.Request) {
	// r.Context()にはotelhttp.NewHandlerで作成されたスパンのコンテキストが含まれる
	// 手動でスパンを作成・終了する必要がない

	// 認証処理は自動計装ではカバーされないため手動計装のまま
	userID, err := authenticate(r.Context(), r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// リクエストのbodyを構造体にデコード
	var req MemoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// コンテキストを渡せば、子スパンとして関連付けられる
	if err := callBackend(r.Context(), userID, req.Content); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, `{"status":"created"}`)
}

func authenticate(ctx context.Context, r *http.Request) (string, error) {
	tracer := otel.Tracer("frontend")
	_, span := tracer.Start(ctx, "authenticate")
	defer span.End()

	auth := r.Header.Get("Authorization")
	if auth == "" {
		err := fmt.Errorf("missing authorization header")
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing authorization header")
		return "", err
	}

	// ダミーの認証処理: "Bearer token-{ユーザーID}"形式のトークンからユーザーIDを取り出す
	userID, ok := strings.CutPrefix(auth, "Bearer token-")
	if !ok || userID == "" {
		err := fmt.Errorf("invalid authorization token")
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid authorization token")
		return "", err
	}

	// 認証成功時にユーザーIDを記録
	span.SetAttributes(attribute.String("user_id", userID))
	span.SetStatus(codes.Ok, "")
	return userID, nil
}

func callBackend(ctx context.Context, userID, content string) error {
	reqBody, err := json.Marshal(InternalMemoRequest{
		UserID:  userID,
		Content: content,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", backendURL+"/internal/memos", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// otelhttp.NewTransportでラップしたクライアントを使用
	// コンテキスト伝播とスパン作成が自動的に行われる
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call backend: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	return nil
}

func main() {
	shutdown, err := initTracer()
	if err != nil {
		slog.Error("failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer shutdown()

	if u := os.Getenv("BACKEND_URL"); u != "" {
		backendURL = u
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /memos", handleCreateMemo)

	// otelhttp.NewHandlerでHTTPサーバーをラップ
	handler := otelhttp.NewHandler(mux, "frontend")

	slog.Info("Frontend server starting on :8080", "backend_url", backendURL)
	if err := http.ListenAndServe(":8080", handler); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
