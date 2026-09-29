package lineworks_usecase

import (
	"strings"
	"testing"
)

func TestDueDigestMessageWithItems(t *testing.T) {
	body := DueDigestMessage(
		[]DueDigestItem{{
			Title: "当日案件", AskerName: "田中", DueDate: "2026-09-30", Link: "https://solvi.example/questions/a",
		}},
		[]DueDigestItem{{
			Title: "明日案件", AskerName: "", DueDate: "2026-10-01", Link: "https://solvi.example/questions/b",
		}},
	)
	if !strings.Contains(body, "【定期通知】") {
		t.Fatalf("body=%s", body)
	}
	if !strings.Contains(body, "期日当日の問い合わせが、1件あります。") {
		t.Fatalf("body=%s", body)
	}
	if !strings.Contains(body, "内容を確認し回答してください。") {
		t.Fatalf("body=%s", body)
	}
	if !strings.Contains(body, "・当日案件 田中 2026-09-30 https://solvi.example/questions/a") {
		t.Fatalf("body=%s", body)
	}
	if !strings.Contains(body, "期日が明日に迫った問い合わせが、1件あります。") {
		t.Fatalf("body=%s", body)
	}
	if !strings.Contains(body, "・明日案件 不明 2026-10-01 https://solvi.example/questions/b") {
		t.Fatalf("body=%s", body)
	}
	if strings.Index(body, "期日当日") > strings.Index(body, "期日が明日") {
		t.Fatalf("today block should come before tomorrow block: %s", body)
	}
}

func TestDueDigestMessageEmpty(t *testing.T) {
	body := DueDigestMessage(nil, nil)
	want := "【定期通知】\n期日に近づいている問い合わせはありませんでした。"
	if body != want {
		t.Fatalf("body=%q want=%q", body, want)
	}
}

func TestDueDigestMessageOmitsEmptyBlock(t *testing.T) {
	body := DueDigestMessage([]DueDigestItem{{
		Title: "当日のみ", AskerName: "佐藤", DueDate: "2026-09-30", Link: "https://solvi.example/questions/c",
	}}, nil)
	if strings.Contains(body, "期日が明日に迫った") {
		t.Fatalf("body=%s", body)
	}
}
