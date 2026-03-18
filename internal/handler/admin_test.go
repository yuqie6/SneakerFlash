package handler

import (
	"testing"
	"time"
)

func TestParseAdminTime_UsesLocalTimeForDatetimeLocal(t *testing.T) {
	got, err := parseAdminTime("2026-03-18T10:00")
	if err != nil {
		t.Fatalf("parseAdminTime() error = %v", err)
	}

	want := time.Date(2026, 3, 18, 10, 0, 0, 0, time.Local)
	if !got.Equal(want) || got.Location() != time.Local {
		t.Fatalf("parseAdminTime() = %v (%v), want %v (%v)", got, got.Location(), want, time.Local)
	}
}
