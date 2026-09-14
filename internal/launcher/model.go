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
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
)

var ErrNoTTY = errors.New("interactive mode requires a terminal; use `hosta list` or `hosta connect <host>`")

var (
	accentStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FB923C"))
	titleStyle    = accentStyle.Bold(true)
	nameStyle     = lipgloss.NewStyle().Bold(true)
	aliasStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#67E8F9"))
	addressStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6EE7B7"))
	metadataStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8"))
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#C2410C")).
			Bold(true)
)

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
	input.Prompt = accentStyle.Render("Search") + "  "
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
	title := titleStyle.Render("Hosta")
	count := metadataStyle.Render(fmt.Sprintf("%d hosts", len(m.results)))
	fmt.Fprintf(&content, "%s%s%s\n\n%s\n\n", title, strings.Repeat(" ", max(1, m.width-lipgloss.Width(title)-lipgloss.Width(count))), count, m.input.View())

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

	content.WriteString("\n" + metadataStyle.Render("↑↓ select  •  Enter connect  •  Esc clear/quit  •  Tab details  •  Ctrl+C quit"))
	view := tea.NewView(content.String())
	view.AltScreen = true
	return view
}

func (m *model) refresh() {
	m.results = m.index.Search(m.input.Value(), 0)
	m.cursor = 0
}

func (m model) visibleRange() (int, int) {
	reserved := 7
	if m.details {
		reserved += 3
	}
	limit := max(1, m.height-reserved)
	start := 0
	if m.cursor >= limit {
		start = m.cursor - limit + 1
	}
	return start, min(len(m.results), start+limit)
}

func (m model) renderResult(content *strings.Builder, position int) {
	value := m.results[position].Host
	if m.width < 40 || m.height < 12 {
		prefix := "  "
		if position == m.cursor {
			prefix = "› "
		}
		fmt.Fprintln(content, m.styleRow(prefix+value.Alias, position == m.cursor))
		return
	}

	label := value.DisplayName
	if label == "" {
		label = value.Alias
	}
	address := formatPreview(value.Preview)
	metadata := hostMetadata(value)
	selected := position == m.cursor
	if selected {
		parts := []string{"›", label}
		if aliases := displayAliases(value); aliases != label {
			parts = append(parts, aliases)
		}
		if address != "" {
			parts = append(parts, address)
		}
		if metadata != "" {
			parts = append(parts, metadata)
		}
		fmt.Fprintln(content, m.styleRow(strings.Join(parts, "  "), true))
	} else {
		parts := []string{"  " + nameStyle.Render(label)}
		if aliases := displayAliases(value); aliases != label {
			parts = append(parts, aliasStyle.Render(aliases))
		}
		if address != "" {
			parts = append(parts, addressStyle.Render(address))
		}
		if metadata != "" {
			parts = append(parts, metadataStyle.Render(metadata))
		}
		fmt.Fprintln(content, ansi.Truncate(strings.Join(parts, "  "), m.width, "…"))
	}
	if selected && m.details {
		if value.Description != "" {
			fmt.Fprintln(content, metadataStyle.Render("    "+value.Description))
		}
		for _, source := range value.Sources {
			fmt.Fprintln(content, metadataStyle.Render(fmt.Sprintf("    %s:%d", source.File, source.Line)))
		}
	}
}

func (m model) styleRow(value string, selected bool) string {
	value = ansi.Truncate(value, max(1, m.width-2), "…")
	if !selected {
		return value
	}
	return selectedStyle.Width(m.width).PaddingLeft(1).Render(value)
}

func hostMetadata(value host.Host) string {
	metadata := append([]string(nil), value.Tags...)
	if value.Group != "" {
		metadata = append([]string{value.Group}, metadata...)
	}
	return strings.Join(metadata, " · ")
}

func displayAliases(value host.Host) string {
	return strings.Join(append([]string{value.Alias}, value.Aliases...), ", ")
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
