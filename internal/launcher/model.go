package launcher

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	input textinput.Model
}

func newModel() model {
	input := textinput.New()
	input.Prompt = "Search  "
	input.Placeholder = "type to filter hosts"
	input.Focus()
	return model{input: input}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.input.Value() == "" {
				return m, tea.Quit
			}
			m.input.SetValue("")
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	title := lipgloss.NewStyle().Bold(true).Render("Hosta")
	content := fmt.Sprintf("%s\n\n%s\n\nSSH Config discovery arrives in M1.\n\nCtrl+C quit", title, m.input.View())
	return tea.NewView(content)
}

func Run(input io.Reader, output io.Writer) error {
	_, err := tea.NewProgram(newModel(), tea.WithInput(input), tea.WithOutput(output)).Run()
	return err
}
