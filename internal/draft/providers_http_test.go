package draft

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnthropicProvider(t *testing.T) {
	var gotAuth, gotModel, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("x-api-key")
		gotPath = r.URL.Path
		gotModel = readJSONField(t, r, "model")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"content":[{"type":"text","text":"---\nnumber: 0\ntitle: Drafted\nstatus: proposed\n---\n\n## Alternatives considered\n- x. Rejected: y.\n"}]}`))
	}))
	defer srv.Close()

	p := NewAnthropic("sk-test", srv.URL)
	out, err := p.Draft(context.Background(), PRData{Number: 1, Title: "t", Diff: "d"})
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "sk-test" || gotPath != "/v1/messages" || gotModel == "" {
		t.Fatalf("request wrong: auth=%q path=%q model=%q", gotAuth, gotPath, gotModel)
	}
	if !strings.Contains(out, "title: Drafted") {
		t.Fatalf("bad output:\n%s", out)
	}
}

func TestOpenAIProvider(t *testing.T) {
	var gotAuth, gotModel, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		gotModel = readJSONField(t, r, "model")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"---\nnumber: 0\ntitle: GPT draft\nstatus: proposed\n---\n"}}]}`))
	}))
	defer srv.Close()

	p := NewOpenAI("sk-oai", srv.URL)
	out, err := p.Draft(context.Background(), PRData{Number: 1, Title: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer sk-oai" || gotPath != "/v1/chat/completions" || gotModel == "" {
		t.Fatalf("request wrong: auth=%q path=%q model=%q", gotAuth, gotPath, gotModel)
	}
	if !strings.Contains(out, "title: GPT draft") {
		t.Fatalf("bad output:\n%s", out)
	}
}

func TestProviderHTTPErrorSurfacesBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer srv.Close()
	p := NewAnthropic("k", srv.URL)
	_, err := p.Draft(context.Background(), PRData{Title: "t"})
	if err == nil || !strings.Contains(err.Error(), "429") || !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("want status+body in error, got %v", err)
	}
}
