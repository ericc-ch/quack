package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type winSize struct {
	width  int
	height int
}

type model struct {
	winSize winSize
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.winSize = winSize{
			width:  msg.Width,
			height: msg.Height,
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	style := lipgloss.
		NewStyle().
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#ff0000")).
		Width(m.winSize.width / 4).
		Height(m.winSize.height / 4)

	style2 := lipgloss.NewStyle().Width(m.winSize.width / 8).Height(m.winSize.height / 8).Inherit(style)

	winSize := fmt.Sprintf("width: %d, height: %d", m.winSize.width, m.winSize.height)
	winSize2 := fmt.Sprintf("width: %d, height: %d", m.winSize.width, m.winSize.height)

	return lipgloss.JoinHorizontal(
		0.8,
		style.Render(winSize),
		style2.Render(winSize2),
	)
}
