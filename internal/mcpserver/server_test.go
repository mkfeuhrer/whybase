package mcpserver

import (
	"context"
	"strings"
	"testing"

	"github.com/mkfeuhrer/whybase/internal/adr"
	"github.com/mkfeuhrer/whybase/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func testIndex(t *testing.T) *store.Index {
	t.Helper()
	add := func(ix *store.Index, n int, title, status, body string, tags []string) {
		r := adr.Record{Number: n, Title: title, Status: adr.Status(status),
			Date: "2026-01-01", Body: body, Tags: tags, Path: "/x"}
		ix.ByNumber[n] = r
		ix.Ordered = append(ix.Ordered, r)
	}
	ix := &store.Index{ByNumber: map[int]adr.Record{}}
	add(ix, 1, "Use Postgres for sessions", "accepted",
		"\n## Context\nsessions.\n\n## Decision\npg.\n\n## Alternatives considered\n- Redis. Rejected: cost.", []string{"storage", "sessions"})
	add(ix, 2, "Kafka for events", "accepted",
		"\n## Context\nvolume.\n\n## Decision\nkafka.\n\n## Alternatives considered\n- SQS. Rejected: ordering.", nil)
	return ix
}

func connect(t *testing.T, srv *mcp.Server) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ss, err := srv.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	cs, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close(); _ = ss.Close() })
	return cs
}

func call(t *testing.T, sess *mcp.ClientSession, name string, args any) string {
	t.Helper()
	res, err := sess.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("tool %s: %v", name, err)
	}
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

func TestServerTools(t *testing.T) {
	srv := New(testIndex(t))
	sess := connect(t, srv)

	out := call(t, sess, "search_decisions", map[string]any{"query": "postgres"})
	if !strings.Contains(out, "Use Postgres") {
		t.Fatalf("search_decisions result missing record:\n%s", out)
	}

	out = call(t, sess, "get_decision", map[string]any{"number": 2})
	if !strings.Contains(out, "Kafka") || !strings.Contains(out, "Alternatives") {
		t.Fatalf("get_decision result wrong:\n%s", out)
	}

	out = call(t, sess, "check_paths", map[string]any{"paths": []string{"internal/sessions/store.go"}})
	if !strings.Contains(out, "Postgres") {
		t.Fatalf("check_paths should surface session ADR:\n%s", out)
	}
}

func TestListToolsExposesThree(t *testing.T) {
	srv := New(testIndex(t))
	sess := connect(t, srv)
	names := map[string]bool{}
	for tool, err := range sess.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		names[tool.Name] = true
	}
	for _, want := range []string{"search_decisions", "get_decision", "check_paths"} {
		if !names[want] {
			t.Fatalf("missing tool %q; got %v", want, names)
		}
	}
}

func TestGetDecisionMissingErrors(t *testing.T) {
	srv := New(testIndex(t))
	sess := connect(t, srv)
	res, err := sess.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "get_decision", Arguments: map[string]any{"number": 42}})
	if err != nil {
		t.Fatalf("transport error not expected: %v", err)
	}
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	if !res.IsError || !strings.Contains(b.String(), "no record 42") {
		t.Fatalf("want IsError result mentioning 'no record 42', got isError=%v text=%q", res.IsError, b.String())
	}
}
