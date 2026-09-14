package launcher

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/HobaiRiku/hosta/internal/host"
)

func TestInitialView(t *testing.T) {
	m := newModel(testIndex(t))
	result := m.View()
	if !result.AltScreen {
		t.Fatal("launcher must use the alternate screen buffer")
	}
	view := result.Content
	for _, want := range []string{"Hosta", "Search", "2 hosts", "Home Server", "root@home.example.com:22"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view %q does not contain %q", view, want)
		}
	}
}

func TestTypingFiltersHosts(t *testing.T) {
	m := newModel(testIndex(t))
	next, _ := m.Update(tea.KeyPressMsg{Code: '路', Text: "路"})
	got := next.(model)
	if len(got.results) != 1 || got.results[0].Host.Alias != "router" {
		t.Fatalf("results = %#v", got.results)
	}
}

func TestNavigationAndSelection(t *testing.T) {
	m := newModel(testIndex(t))
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = next.(model)
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
	next, command := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = next.(model)
	if m.selected != "router" || command == nil {
		t.Fatalf("selected = %q, command = %v", m.selected, command)
	}
}

func TestEscapeClearsThenQuits(t *testing.T) {
	m := newModel(testIndex(t))
	m.input.SetValue("prod")
	m.refresh()
	next, command := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(model)
	if m.input.Value() != "" || command != nil {
		t.Fatalf("query = %q, command = %v", m.input.Value(), command)
	}
	_, command = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if command == nil {
		t.Fatal("second escape did not quit")
	}
}

func TestSmallTerminalUsesCompactRows(t *testing.T) {
	m := newModel(testIndex(t))
	next, _ := m.Update(tea.WindowSizeMsg{Width: 30, Height: 10})
	view := next.(model).View().Content
	if !strings.Contains(view, "> home") || strings.Contains(view, "root@home.example.com") {
		t.Fatalf("compact view = %q", view)
	}
}

func TestTabShowsDetails(t *testing.T) {
	index, err := host.NewIndex([]host.Host{{
		Alias:       "home",
		Description: "Primary host",
		Sources:     []host.Source{{File: "/tmp/config", Line: 4}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	m := newModel(index)
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	view := next.(model).View().Content
	if !strings.Contains(view, "Primary host") || !strings.Contains(view, "/tmp/config:4") {
		t.Fatalf("details view = %q", view)
	}
}

func TestEmptyIndexMessage(t *testing.T) {
	index, err := host.NewIndex(nil)
	if err != nil {
		t.Fatal(err)
	}
	if view := newModel(index).View().Content; !strings.Contains(view, "No SSH hosts discovered") {
		t.Fatalf("empty view = %q", view)
	}
}

func TestRunRejectsNonTTY(t *testing.T) {
	_, err := Run(&bytes.Buffer{}, &bytes.Buffer{}, testIndex(t))
	if !errors.Is(err, ErrNoTTY) {
		t.Fatalf("Run() error = %v, want ErrNoTTY", err)
	}
}

func testIndex(t *testing.T) *host.Index {
	t.Helper()
	index, err := host.NewIndex([]host.Host{
		{Alias: "home", DisplayName: "Home Server", Group: "personal", Tags: []string{"home"}, Preview: host.Preview{HostName: "home.example.com", User: "root", Port: "22"}},
		{Alias: "router", DisplayName: "路由器", Group: "personal", Tags: []string{"network"}, Preview: host.Preview{HostName: "192.0.2.1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return index
}

func BenchmarkFilter500Hosts(b *testing.B) {
	hosts := make([]host.Host, 500)
	for i := range hosts {
		hosts[i] = host.Host{Alias: fmt.Sprintf("production-%03d", i), DisplayName: "Production Host"}
	}
	index, err := host.NewIndex(hosts)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for range b.N {
		m := newModel(index)
		_, _ = m.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	}
}
