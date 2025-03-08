package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const url = "https://charm.sh/"

type model struct {
	status int
	err    error
}

func checkServer() tea.Msg {
	c := &http.Client{Timeout: time.Second}
	res, err := c.Get(url)

	if err != nil {
		return errMsg{err}
	}

	return statusMsg(res.StatusCode)
}

type statusMsg int

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

func (m model) Init() tea.Cmd {
	return checkServer
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("we had some trouble: %v\n\n", m.err)
	}

	s := fmt.Sprintf("Checking %s ...", url)

	if m.status > 0 {
		s += fmt.Sprintf("%d %s!", m.status, http.StatusText(m.status))
	}

	return s + "\n"
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case statusMsg:
		m.status = int(msg)
		return m, tea.Quit

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	return m, nil
}

func main() {
	p := tea.NewProgram(model{})

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
