package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type action int

const (
	actionSkipped action = iota
	actionSuccess
	actionError
)

var sedEscapeChars = []string{"/", "\"", "[", "]", "(", ")", "{", "}"}

type model struct {
	index   int
	prompts [][2]string
	input   textinput.Model
	action  action
	toast   string
}

func newModel(prompts [][2]string) model {
	cmdInput := textinput.New()
	cmdInput.ShowSuggestions = true
	cmdInput.Focus()
	cmdInput.CharLimit = 512
	cmdInput.Width = 512

	return model{
		index:   0,
		prompts: prompts,
		input:   cmdInput,
		toast:   "\n",
	}
}

func (m model) Init() tea.Cmd {
	return tea.Sequence(textinput.Blink, func() tea.Msg { return 0 })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case int:
		m.index = msg

		if m.index >= len(m.prompts) {
			return m, tea.Quit
		}

		m.input.Prompt = fmt.Sprintf("[%d/%d] ", m.index+1, len(m.prompts)) + filenameStyle.Render(m.prompts[m.index][0]) + "\n> "
		m.input.SetValue(m.prompts[m.index][1])
		m.input.CursorEnd()

		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			find := escapeChars(m.prompts[m.index][1], sedEscapeChars...)
			replace := escapeChars(m.input.Value(), sedEscapeChars...)

			if find == replace {
				m.action = actionSkipped
				m.toast = "Skipped " + m.prompts[m.index][0]

				return m, func() tea.Msg { return m.index + 1 }
			}

			sedQuery := fmt.Sprintf("s/%s/%s/g", find, replace)
			execCmd := exec.Command("sed", "-i", sedQuery, m.prompts[m.index][0])

			err := execCmd.Run()

			if err != nil {
				m.action = actionError
				m.toast = "sed: " + err.Error()
			} else {
				m.action = actionSuccess
				m.toast = "Updated " + m.prompts[m.index][0]
			}

			return m, func() tea.Msg { return m.index + 1 }
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	output := m.input.View() + "\n"

	if len(m.toast) > 0 {
		switch m.action {
		case actionSkipped:
			output += "\n" + filenameStyle.Render(m.toast)
		case actionSuccess:
			output += "\n" + successStyle.Render(m.toast)
		case actionError:
			output += "\n" + errorStyle.Render(m.toast)
		}
	}

	return output
}

func escapeChars(s string, chars ...string) string {
	for _, char := range chars {
		s = strings.ReplaceAll(s, char, "\\"+char)
	}

	return s
}
