package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mkfeuhrer/whybase/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New builds an MCP server exposing the decision index to AI coding agents.
func New(ix *store.Index) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{Name: "whybase", Version: "0.1.0"}, nil)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "search_decisions",
		Description: "Search architecture decision records by keywords. Returns number, title, status, date.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct {
		Query string `json:"query" jsonschema:"keywords to search for"`
	}) (*mcp.CallToolResult, any, error) {
		hits := ix.Search(in.Query)
		type row struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			Status string `json:"status"`
			Date   string `json:"date"`
		}
		rows := make([]row, len(hits))
		for i, r := range hits {
			rows[i] = row{r.Number, r.Title, string(r.Status), r.Date}
		}
		if len(rows) == 0 {
			return textResult("no matching decisions"), nil, nil
		}
		return jsonResult(rows)
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_decision",
		Description: "Get one full architecture decision record by number.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct {
		Number int `json:"number" jsonschema:"ADR number"`
	}) (*mcp.CallToolResult, any, error) {
		r, ok := ix.Get(in.Number)
		if !ok {
			return nil, nil, fmt.Errorf("no record %d", in.Number)
		}
		return jsonResult(struct {
			Number       int      `json:"number"`
			Title        string   `json:"title"`
			Status       string   `json:"status"`
			Date         string   `json:"date"`
			Supersedes   []int    `json:"supersedes,omitempty"`
			SupersededBy []int    `json:"superseded_by,omitempty"`
			Tags         []string `json:"tags,omitempty"`
			Body         string   `json:"body"`
		}{r.Number, r.Title, string(r.Status), r.Date, r.Supersedes, r.SupersededBy, r.Tags, r.Body})
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "check_paths",
		Description: "Given file paths you are about to edit, returns governing architecture decisions so you do not violate or re-propose rejected approaches.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct {
		Paths []string `json:"paths" jsonschema:"file paths about to be modified"`
	}) (*mcp.CallToolResult, any, error) {
		hits := ix.Governing(in.Paths)
		var b strings.Builder
		if len(hits) == 0 {
			b.WriteString("No recorded decisions govern these paths.")
		}
		for _, r := range hits {
			fmt.Fprintf(&b, "ADR-%04d [%s] %s\n", r.Number, r.Status, r.Title)
			for _, alt := range alternativesLines(r.Body) {
				fmt.Fprintf(&b, "  rejected: %s\n", alt)
			}
		}
		return textResult(b.String()), nil, nil
	})

	return srv
}

// Run serves MCP over stdio (used by `whybase mcp`).
func Run(ctx context.Context, ix *store.Index) error {
	return New(ix).Run(ctx, &mcp.StdioTransport{})
}

// alternativesLines extracts the bullet lines under "## Alternatives considered".
func alternativesLines(body string) []string {
	var out []string
	in := false
	for _, line := range strings.Split(body, "\n") {
		t := strings.TrimSpace(line)
		low := strings.ToLower(t)
		if strings.HasPrefix(low, "## ") {
			in = strings.HasPrefix(low, "## alternatives")
			continue
		}
		if in && strings.HasPrefix(t, "-") {
			out = append(out, strings.TrimPrefix(t, "- "))
		}
	}
	return out
}

func textResult(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}}
}

func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return textResult(string(b)), nil, nil
}
