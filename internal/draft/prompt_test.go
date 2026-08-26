package draft

import (
	"strings"
	"testing"
)

func TestBuildPromptRequiresAlternatives(t *testing.T) {
	p := PRData{Number: 42, Title: "Switch cache to Redis", Body: "Latency.", Diff: "diff --git a/cache.go"}
	prompt, err := BuildPrompt(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Alternatives considered",
		"proposed",
		"front-matter",
		"number:",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
}

func TestBuildPromptTruncatesHugeDiff(t *testing.T) {
	big := strings.Repeat("x", 100_000)
	prompt, err := BuildPrompt(PRData{Number: 1, Title: "t", Diff: big})
	if err != nil {
		t.Fatal(err)
	}
	if len(prompt) > 70_000 {
		t.Fatalf("prompt not truncated: %d chars", len(prompt))
	}
	if !strings.Contains(prompt, "[diff truncated") {
		t.Fatal("truncation not marked")
	}
}
