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

var backendURL = "http://localhost:8081"

// バックエンド呼び出しに使用するHTTPクライアント
var httpClient = &http.Client{}

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
	// HTTPヘッダーからトレースコンテキストを抽出
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	tracer := otel.Tracer("frontend")
	ctx, span := tracer.Start(ctx, "handleCreateMemo",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	// リクエストの属性を記録
	span.SetAttributes(
		semconv.HTTPRequestMethodKey.String(r.Method),
		semconv.URLPath(r.URL.Path),
	)

	userID, err := authenticate(ctx, r)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "authentication failed")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	span.SetAttributes(attribute.String("user_id", userID))

	// リクエストのbodyを構造体にデコード
	var req MemoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decode request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// ctxを渡すことで親子関係が構築される
	if err := callBackend(ctx, userID, req.Content); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to call backend")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// 正常終了時はOKステータスを設定
	span.SetStatus(codes.Ok, "")
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
	tracer := otel.Tracer("frontend")
	ctx, span := tracer.Start(ctx, "callBackend",
		trace.WithSpanKind(trace.SpanKindClient), // リモート呼び出しはクライアントとして扱う
	)
	defer span.End()

	// バックエンド呼び出しの属性を記録
	span.SetAttributes(
		semconv.HTTPRequestMethodKey.String("POST"),
		attribute.String("peer.service", "backend"),
	)

	reqBody, err := json.Marshal(InternalMemoRequest{
		UserID:  userID,
		Content: content,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal request body")
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", backendURL+"/internal/memos", bytes.NewReader(reqBody))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create request")
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// トレースコンテキストをHTTPヘッダーに注入
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to call backend")
		return fmt.Errorf("failed to call backend: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("backend returned status %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, "backend returned error")
		return err
	}

	span.SetStatus(codes.Ok, "")
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

	slog.Info("Frontend server starting on :8080", "backend_url", backendURL)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
