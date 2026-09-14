package launcher

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestInitialView(t *testing.T) {
	view := newModel().View().Content
	for _, want := range []string{"Hosta", "Search", "SSH Config discovery arrives in M1."} {
		if !strings.Contains(view, want) {
			t.Fatalf("view %q does not contain %q", view, want)
		}
	}
}

func TestEscapeClearsQuery(t *testing.T) {
	m := newModel()
	m.input.SetValue("prod")
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got := next.(model)
	if got.input.Value() != "" {
		t.Fatalf("query = %q, want empty", got.input.Value())
	}
}
