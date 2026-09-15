package datetime

import (
	"testing"
	"time"
)

func TestParseAnswerDueDate_EndOfDayJST(t *testing.T) {
	got, err := ParseAnswerDueDate("2026-09-14")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 9, 14, 23, 59, 59, 0, jst)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParseAnswerDueDate_Invalid(t *testing.T) {
	_, err := ParseAnswerDueDate("invalid")
	if err == nil {
		t.Fatal("expected error")
	}
}
