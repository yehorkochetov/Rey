package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/yehorkochetov/rey/internal/scanner"
)

// captureStdout redirects os.Stdout for the duration of fn and returns
// everything written. RenderGraveyard writes directly to stdout, so tests
// must intercept the fd rather than inspect a return value.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()

	_ = w.Close()
	os.Stdout = orig
	<-done
	return buf.String()
}

func TestRenderGraveyard_EmptyShowsPlaceholder(t *testing.T) {
	out := captureStdout(t, func() {
		RenderGraveyard(nil)
	})
	if !strings.Contains(out, "No wasted resources found") {
		t.Errorf("expected empty-state placeholder, got: %q", out)
	}
}

func TestRenderGraveyard_SingleRendersWithoutPanic(t *testing.T) {
	res := []scanner.DeadResource{
		{Type: "EIP", ID: "eipalloc-1", Region: "us-east-1", MonthlyCost: 3.60},
	}
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("render panicked: %v", p)
		}
	}()
	out := captureStdout(t, func() { RenderGraveyard(res) })
	if out == "" {
		t.Error("expected non-empty output for a single resource")
	}
}

func TestRenderGraveyard_MultipleRendersWithoutPanic(t *testing.T) {
	res := []scanner.DeadResource{
		{Type: "EIP", ID: "eipalloc-1", Region: "us-east-1", MonthlyCost: 3.60},
		{Type: "EBSVolume", ID: "vol-2", Region: "us-east-1", MonthlyCost: 10.00},
		{Type: "NATGateway", ID: "nat-3", Region: "us-east-1", MonthlyCost: 32.40},
	}
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("render panicked: %v", p)
		}
	}()
	_ = captureStdout(t, func() { RenderGraveyard(res) })
}

func TestRenderGraveyard_TotalCostSums(t *testing.T) {
	// $3.60 + $10.00 = $13.60 — the footer string is load-bearing.
	res := []scanner.DeadResource{
		{Type: "EIP", ID: "eipalloc-1", Region: "us-east-1", MonthlyCost: 3.60},
		{Type: "EBSVolume", ID: "vol-2", Region: "us-east-1", MonthlyCost: 10.00},
	}
	out := captureStdout(t, func() { RenderGraveyard(res) })
	if !strings.Contains(out, "$13.60") {
		t.Errorf("expected total $13.60 in footer, got: %q", out)
	}
	if !strings.Contains(out, "2 resources") {
		t.Errorf("expected footer to mention 2 resources, got: %q", out)
	}
}

func TestCostStyle(t *testing.T) {
	// lipgloss.TerminalColor doesn't offer equality, but the underlying
	// lipgloss.Color is a string — fmt.Sprintf gives a stable comparable
	// key across all TerminalColor implementations.
	fgOf := func(cost float64) string { return fgKey(costStyle(cost)) }

	cases := []struct {
		name string
		cost float64
		want string
	}{
		{"10.01 gets red (high cost)", 10.01, fgKey(highCost)},
		{"1.01 gets yellow (medium cost)", 1.01, fgKey(medCost)},
		{"0.50 gets default style", 0.50, fgKey(defaultCost)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := fgOf(c.cost); got != c.want {
				t.Errorf("costStyle(%v) foreground = %q, want %q", c.cost, got, c.want)
			}
		})
	}
}

func TestCostStyle_Boundaries(t *testing.T) {
	// Boundaries in costStyle are strictly ">": 10 lands in the medium
	// bucket, 1 lands in the default bucket. Pin this so a stray >=
	// can't silently shift colors for cut-off-priced resources.
	if fgKey(costStyle(10)) == fgKey(highCost) {
		t.Error("cost == 10 should NOT be red (boundary is strictly > 10)")
	}
	if fgKey(costStyle(1)) == fgKey(medCost) {
		t.Error("cost == 1 should NOT be yellow (boundary is strictly > 1)")
	}
}

// fgKey returns a comparable string key for a style's foreground color.
// Lipgloss's TerminalColor interface isn't directly comparable, but the
// concrete Color/NoColor types print distinctly under %#v — good enough
// to tell "which style did costStyle return" in a test.
func fgKey(s lipgloss.Style) string {
	return fmt.Sprintf("%#v", s.GetForeground())
}
