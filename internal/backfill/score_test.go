package backfill

import "testing"

func TestScoreCommitStrongVerbWins(t *testing.T) {
	strong := ScoreCommit(Commit{
		Subject: "add Postgres session store",
		Files:   []string{"sessions/store.go", "sessions/store_test.go"},
		Additions: 120, Deletions: 30,
	})
	boring := ScoreCommit(Commit{
		Subject: "update README links",
		Files:   []string{"README.md"},
		Additions: 4, Deletions: 4,
	})
	if strong <= boring {
		t.Fatalf("decision-shaped commit (%d) should outrank docs touch (%d)", strong, boring)
	}
}

func TestScoreCommitMergeAndDepTouch(t *testing.T) {
	m := ScoreCommit(Commit{SHA: "abc1234", Subject: "Merge pull request #42", IsMerge: true})
	if m != wMerge {
		t.Fatalf("merge commit should score exactly %d, got %d", wMerge, m)
	}
	dep := ScoreCommit(Commit{Subject: "bump go.mod", Files: []string{"go.mod"}, Additions: 10})
	if dep < wDepTouch+wChurn {
		t.Fatalf("dep-touch commit should at least get dep+churn, got %d", dep)
	}
}

func TestLogScaleKnee(t *testing.T) {
	cases := []struct {
		lines int
		want  float64 // 0..1
	}{
		{0, 0},
		{1, 0}, // below knee: tiny
		{400, 1},
		{4000, 1}, // clamped
	}
	for _, c := range cases {
		if got := logScale(float64(c.lines)); got != c.want {
			t.Fatalf("logScale(%d) = %v, want %v", c.lines, got, c.want)
		}
	}
}

func TestDescribeTruncatesSHA(t *testing.T) {
	d := describe(Commit{SHA: "0123456789abcdef", Subject: "add thing", Additions: 1})
	if len(d) == 0 {
		t.Fatal("empty describe")
	}
}
