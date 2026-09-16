package logutils_test

import (
	"log/slog"
	"net/http"
	logutils "solvi/internal/shared/logUtils"
	"testing"
	"unsafe"
)

type validStruct struct {
	ID   int
	Name string
}

type funcStruct struct {
	Fn func()
}

type chanStruct struct {
	Ch chan int
}

type unsafeStruct struct {
	Ptr unsafe.Pointer
}

type nestedStruct struct {
	Data funcStruct
}

func TestLogAny(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		v        any
		wantKind slog.Kind
	}{
		{
			name:     "struct",
			key:      "value",
			v:        validStruct{ID: 1, Name: "test"},
			wantKind: slog.KindAny,
		},
		{
			name:     "struct pointer",
			key:      "value",
			v:        &validStruct{ID: 1, Name: "test"},
			wantKind: slog.KindAny,
		},
		{
			name:     "nil",
			key:      "value",
			v:        nil,
			wantKind: slog.KindAny,
		},
		{
			name:     "func field",
			key:      "value",
			v:        funcStruct{},
			wantKind: slog.KindString,
		},
		{
			name:     "chan field",
			key:      "value",
			v:        chanStruct{},
			wantKind: slog.KindString,
		},
		{
			name:     "unsafe pointer field",
			key:      "value",
			v:        unsafeStruct{},
			wantKind: slog.KindString,
		},
		{
			name:     "nested unsupported field",
			key:      "value",
			v:        nestedStruct{},
			wantKind: slog.KindString,
		},
		{
			name:     "slice",
			key:      "value",
			v:        []validStruct{},
			wantKind: slog.KindAny,
		},
		{
			name:     "slice contains func",
			key:      "value",
			v:        []funcStruct{},
			wantKind: slog.KindString,
		},
		{
			name:     "map",
			key:      "value",
			v:        map[string]validStruct{},
			wantKind: slog.KindAny,
		},
		{
			name:     "map contains func",
			key:      "value",
			v:        map[string]funcStruct{},
			wantKind: slog.KindString,
		},
		{
			name:     "raw func",
			key:      "value",
			v:        func() {},
			wantKind: slog.KindString,
		},
		{
			name:     "raw chan",
			key:      "value",
			v:        make(chan int),
			wantKind: slog.KindString,
		},
		{
			name:     "http request",
			key:      "value",
			v:        &http.Request{},
			wantKind: slog.KindString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := logutils.LogAny(tt.key, tt.v)

			if got.Key != tt.key {
				t.Fatalf("key mismatch: got=%s want=%s", got.Key, tt.key)
			}

			if got.Value.Kind() != tt.wantKind {
				t.Fatalf(
					"kind mismatch: got=%v want=%v",
					got.Value.Kind(),
					tt.wantKind,
				)
			}
		})
	}

	t.Run("http request should output type name", func(t *testing.T) {
		got := logutils.LogAny("request", &http.Request{})

		if got.Value.Kind() != slog.KindString {
			t.Fatalf("expected KindString")
		}

		if got.Value.String() != "<*http.Request>" {
			t.Fatalf(
				"unexpected value: got=%s",
				got.Value.String(),
			)
		}
	})
}
