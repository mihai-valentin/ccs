package summarizer

import "testing"

func makeMessages(n int) []message {
	m := make([]message, n)
	for i := 0; i < n; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		m[i] = message{Role: role, Content: string(rune('a' + i%26))}
	}
	return m
}

func TestBuildExcerpt_ShortConversationPassThrough(t *testing.T) {
	// n <= headMessages+tailMessages (5+5) returns the whole slice unchanged.
	in := makeMessages(10)
	got := buildExcerpt(in)
	if len(got) != len(in) {
		t.Fatalf("expected pass-through, got %d messages vs %d", len(got), len(in))
	}
	for i := range got {
		if got[i] != in[i] {
			t.Errorf("message %d differs: got %+v want %+v", i, got[i], in[i])
		}
	}
}

func TestBuildExcerpt_LongConversationIncludesHeadMidTail(t *testing.T) {
	n := 100
	in := makeMessages(n)
	got := buildExcerpt(in)

	// Expected: head(5) + mid(3) + tail(5) = 13
	if len(got) != headMessages+midMessages+tailMessages {
		t.Errorf("expected %d messages, got %d", headMessages+midMessages+tailMessages, len(got))
	}
	// Head must be the first 5.
	for i := 0; i < headMessages; i++ {
		if got[i] != in[i] {
			t.Errorf("head[%d] got %+v want %+v", i, got[i], in[i])
		}
	}
	// Tail must be the last 5.
	for i := 0; i < tailMessages; i++ {
		if got[len(got)-tailMessages+i] != in[n-tailMessages+i] {
			t.Errorf("tail[%d] mismatch", i)
		}
	}
}

func TestBuildExcerpt_NoOverlapBetweenMidAndTail(t *testing.T) {
	// Edge case: n just big enough to trigger mid but tail would overlap.
	// headMessages(5) + midMessages(3) + tailMessages(5) = 13
	// With n=14, midStart=7, mid slice covers [7,8,9], tail would start at 9 → overlap.
	// The fix in buildExcerpt clamps tailStart to midStart+midMessages.
	in := makeMessages(14)
	got := buildExcerpt(in)

	seen := make(map[int]bool)
	for _, m := range got {
		for i := range in {
			if in[i] == m && !seen[i] {
				seen[i] = true
				break
			}
		}
	}
	if len(seen) != len(got) {
		t.Errorf("excerpt contains duplicate messages: %d unique vs %d total", len(seen), len(got))
	}
}

func TestBuildPrompt_IncludesConversation(t *testing.T) {
	msgs := []message{
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "hello"},
	}
	out := buildPrompt(msgs)
	for _, want := range []string{"Summarize", "[user]: hi", "[assistant]: hello", "Summary:"} {
		if !contains(out, want) {
			t.Errorf("prompt missing %q\n---\n%s", want, out)
		}
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
