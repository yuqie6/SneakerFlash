package middlerware

import "testing"

func TestLocalLimiter_Allow(t *testing.T) {
	limiter := NewLocalLimiter(1, 1)

	if !limiter.Allow("user:1") {
		t.Fatal("first Allow() = false, want true")
	}
	if limiter.Allow("user:1") {
		t.Fatal("second Allow() = true, want false")
	}
	if !limiter.Allow("user:2") {
		t.Fatal("different key should have independent bucket")
	}
}

func TestStatusFamily(t *testing.T) {
	tests := map[int]string{
		101: "1xx",
		204: "2xx",
		302: "3xx",
		404: "4xx",
		503: "5xx",
	}

	for code, want := range tests {
		if got := statusFamily(code); got != want {
			t.Fatalf("statusFamily(%d) = %q, want %q", code, got, want)
		}
	}
}
