package scanner

import (
	"testing"
	"time"
)

func TestFormatAge(t *testing.T) {
	cases := []struct {
		name string
		age  time.Duration
		want string
	}{
		{"7 days", 7 * 24 * time.Hour, "7 days"},
		{"1 day", 24 * time.Hour, "1 day"},
		{"exactly 2 days", 48 * time.Hour, "2 days"},
		{"12 hours", 12 * time.Hour, "12 hours"},
		{"1 hour", 1 * time.Hour, "1 hour"},
		{"0 duration renders as dash", 0, "-"},
		{"negative duration renders as dash", -1 * time.Hour, "-"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := DeadResource{Age: c.age}
			got := r.FormatAge()
			if got != c.want {
				t.Errorf("FormatAge(%v) = %q, want %q", c.age, got, c.want)
			}
		})
	}
}

// Guard against a regression where Go's default time.Duration String() leaks
// into the UI — the renderer must not show "168h0m0s" for a week-old resource.
func TestFormatAge_DoesNotLeakDurationString(t *testing.T) {
	r := DeadResource{Age: 7 * 24 * time.Hour}
	got := r.FormatAge()
	if got == (7 * 24 * time.Hour).String() {
		t.Errorf("FormatAge fell through to time.Duration.String(): %q", got)
	}
}

func TestFormatCost(t *testing.T) {
	cases := []struct {
		name string
		cost float64
		want string
	}{
		{"3.6 gets two decimals", 3.6, "$3.60"},
		{"zero", 0, "$0.00"},
		{"integer", 10, "$10.00"},
		{"rounds to two decimals", 1.239, "$1.24"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := DeadResource{MonthlyCost: c.cost}
			got := r.FormatCost()
			if got != c.want {
				t.Errorf("FormatCost(%v) = %q, want %q", c.cost, got, c.want)
			}
		})
	}
}

// A nil Tags map must survive normal read access. Go's spec guarantees this,
// but scanners construct DeadResource in several places and it's worth
// pinning the invariant so a careless refactor can't regress it.
func TestDeadResource_NilTagsDoesNotPanic(t *testing.T) {
	r := DeadResource{}
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("reading nil Tags panicked: %v", p)
		}
	}()
	_ = r.Tags["anything"]
	if _, ok := r.Tags["Name"]; ok {
		t.Errorf("nil map unexpectedly reported key present")
	}
	if n := len(r.Tags); n != 0 {
		t.Errorf("len(nil tags) = %d, want 0", n)
	}
}

func TestDeadResource_EmptyFieldsFormatWithoutPanic(t *testing.T) {
	r := DeadResource{}
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("formatting empty DeadResource panicked: %v", p)
		}
	}()
	if got := r.FormatAge(); got != "-" {
		t.Errorf("empty FormatAge = %q, want %q", got, "-")
	}
	if got := r.FormatCost(); got != "$0.00" {
		t.Errorf("empty FormatCost = %q, want %q", got, "$0.00")
	}
}
