package audit

import "testing"

func TestNormalizeTargetSearchLimit(t *testing.T) {
	if got := normalizeTargetSearchLimit(0); got != defaultTargetSearchLimit {
		t.Fatalf("expected default limit, got %d", got)
	}
	if got := normalizeTargetSearchLimit(maxTargetSearchLimit + 1); got != maxTargetSearchLimit {
		t.Fatalf("expected max limit, got %d", got)
	}
	if got := normalizeTargetSearchLimit(7); got != 7 {
		t.Fatalf("expected explicit limit, got %d", got)
	}
}

func TestLikeKeywordEscapesPercent(t *testing.T) {
	if got := likeKeyword("a%b"); got != "%a\\%b%" {
		t.Fatalf("unexpected escaped keyword: %q", got)
	}
}
