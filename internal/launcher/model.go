package launcher

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/HobaiRiku/hosta/internal/host"
	"github.com/charmbracelet/x/term"
)

var ErrNoTTY = errors.New("interactive mode requires a terminal; use `hosta list` or `hosta connect <host>`")

type model struct {
	input    textinput.Model
	index    *host.Index
	results  []host.Result
	cursor   int
	width    int
	height   int
	details  bool
	selected string
}

func newModel(index *host.Index) model {
	input := textinput.New()
	input.Prompt = "Search  "
	input.Placeholder = "type to filter hosts"
	input.Focus()
	m := model{input: input, index: index, width: 80, height: 24}
	m.refresh()
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.SetWidth(max(10, msg.Width-10))
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.input.Value() == "" {
				return m, tea.Quit
			}
			m.input.SetValue("")
			m.refresh()
			return m, nil
		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "ctrl+n":
			if m.cursor+1 < len(m.results) {
				m.cursor++
			}
			return m, nil
		case "enter":
			if len(m.results) > 0 {
				m.selected = m.results[m.cursor].Host.Alias
				return m, tea.Quit
			}
			return m, nil
		case "tab":
			m.details = !m.details
			return m, nil
		}
	}

	previous := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != previous {
		m.refresh()
	}
	return m, cmd
}

func (m model) View() tea.View {
	var content strings.Builder
	title := lipgloss.NewStyle().Bold(true).Render("Hosta")
	fmt.Fprintf(&content, "%s%*s\n\n%s\n\n", title, max(1, m.width-lipgloss.Width(title)-12), fmt.Sprintf("%d hosts", m.index.Len()), m.input.View())

	if len(m.results) == 0 {
		if m.index.Len() == 0 {
			content.WriteString("No SSH hosts discovered.\n")
		} else {
			content.WriteString("No matching SSH hosts.\n")
		}
	} else {
		start, end := m.visibleRange()
		for position := start; position < end; position++ {
			m.renderResult(&content, position)
		}
	}

	content.WriteString("\n↑↓ select   Enter connect   Esc clear/quit   Tab details   Ctrl+C quit")
	view := tea.NewView(content.String())
	view.AltScreen = true
	return view
}

func (m *model) refresh() {
	m.results = m.index.Search(m.input.Value(), 0)
	m.cursor = 0
}

func (m model) visibleRange() (int, int) {
	limit := max(1, min(20, m.height-8))
	start := 0
	if m.cursor >= limit {
		start = m.cursor - limit + 1
	}
	return start, min(len(m.results), start+limit)
}

func (m model) renderResult(content *strings.Builder, position int) {
	value := m.results[position].Host
	prefix := "  "
	if position == m.cursor {
		prefix = "> "
	}
	if m.width < 40 || m.height < 12 {
		fmt.Fprintf(content, "%s%s\n", prefix, value.Alias)
		return
	}

	label := value.DisplayName
	if label == "" {
		label = value.Alias
	}
	if label == value.Alias {
		fmt.Fprintf(content, "%s%s\n", prefix, label)
	} else {
		fmt.Fprintf(content, "%s%s  %s\n", prefix, label, value.Alias)
	}
	if position != m.cursor {
		return
	}
	metadata := append([]string(nil), value.Tags...)
	if value.Group != "" {
		metadata = append([]string{value.Group}, metadata...)
	}
	if len(metadata) > 0 {
		fmt.Fprintf(content, "    %s\n", strings.Join(metadata, " · "))
	}
	if preview := formatPreview(value.Preview); preview != "" {
		fmt.Fprintf(content, "    %s\n", preview)
	}
	if m.details {
		if value.Description != "" {
			fmt.Fprintf(content, "    %s\n", value.Description)
		}
		for _, source := range value.Sources {
			fmt.Fprintf(content, "    %s:%d\n", source.File, source.Line)
		}
	}
}

func formatPreview(preview host.Preview) string {
	address := preview.HostName
	if address == "" {
		return ""
	}
	if preview.User != "" {
		address = preview.User + "@" + address
	}
	if preview.Port != "" {
		address += ":" + preview.Port
	}
	return address
}

func Run(input io.Reader, output io.Writer, index *host.Index) (string, error) {
	if index == nil {
		return "", fmt.Errorf("host index is nil")
	}
	inputFile, inputOK := input.(*os.File)
	outputFile, outputOK := output.(*os.File)
	if !inputOK || !outputOK || !term.IsTerminal(inputFile.Fd()) || !term.IsTerminal(outputFile.Fd()) {
		return "", ErrNoTTY
	}
	final, err := tea.NewProgram(newModel(index), tea.WithInput(input), tea.WithOutput(output)).Run()
	if err != nil {
		return "", err
	}
	result, ok := final.(model)
	if !ok {
		return "", fmt.Errorf("launcher returned unexpected model %T", final)
	}
	return result.selected, nil
}
