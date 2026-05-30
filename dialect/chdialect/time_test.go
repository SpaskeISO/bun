package chdialect

import (
	"testing"
	"time"
)

func TestDialect_AppendTime(t *testing.T) {
	tm := time.Date(2026, 5, 30, 19, 52, 58, 123456789, time.UTC)

	tests := []struct {
		name   string
		format DateTimeInputFormat
		want   string
	}{
		{
			name:   "basic",
			format: DateTimeInputBasic,
			want:   "'2026-05-30 19:52:58'",
		},
		{
			name:   "best_effort",
			format: DateTimeInputBestEffort,
			want:   "'2026-05-30T19:52:58.123456789Z'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New(WithDateTimeInputFormat(tt.format))
			got := string(d.AppendTime(nil, tm))
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
