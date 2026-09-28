package lineworks_usecase

import (
	"strings"
	"testing"
)

func TestMessagesIncludeQuestionUUIDAndLink(t *testing.T) {
	id := "11111111-1111-1111-1111-111111111111"
	link := QuestionLink("https://solvi.example/", id)
	if link != "https://solvi.example/questions/"+id {
		t.Fatalf("link=%s", link)
	}

	received := ReceivedMessage(id, "田中", "件名", "要約", link)
	answered := AnsweredMessage(id, link)
	reopened := ReopenedMessage(id, "追加です", "佐藤", link)
	for _, body := range []string{received, answered, reopened} {
		if !strings.Contains(body, "#"+id) {
			t.Fatalf("missing uuid: %s", body)
		}
		if !strings.HasSuffix(body, link) {
			t.Fatalf("link is not at the end: %s", body)
		}
		if strings.Contains(body, "#1\n") || strings.Contains(body, "#1】") {
			t.Fatalf("numeric id leaked: %s", body)
		}
	}
	if !strings.HasPrefix(received, "【新規受付 #") {
		t.Fatal(received)
	}
	if !strings.HasPrefix(answered, "【回答完了 #") {
		t.Fatal(answered)
	}
	if !strings.HasPrefix(reopened, "【再対応依頼 #") {
		t.Fatal(reopened)
	}
}

func TestSummarizeTruncatesRunes(t *testing.T) {
	long := strings.Repeat("あ", 201)
	got := Summarize(long)
	if !strings.HasSuffix(got, "…") {
		t.Fatal(got)
	}
	if len([]rune(strings.TrimSuffix(got, "…"))) != 200 {
		t.Fatalf("len=%d", len([]rune(got)))
	}
}
