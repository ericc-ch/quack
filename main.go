package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message represents a chat message
type Message struct {
	content string
	isUser  bool
}

// Model represents the application state
type Model struct {
	messages    []Message
	viewport    viewport.Model
	textarea    textarea.Model
	ready       bool
	senderStyle lipgloss.Style
	systemStyle lipgloss.Style
}

// Initialize the model
func initialModel() Model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.CharLimit = 280
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	return Model{
		messages:    []Message{},
		textarea:    ta,
		viewport:    viewport.New(0, 0),
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		systemStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	// Handle window resize
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// First time sizing
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.messageView())
			m.ready = true
		} else {
			// Subsequent resizes
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
			m.viewport.SetContent(m.messageView())
		}
	}

	// Handle key events
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if key.Alt || key.Alt {
				// Insert a new line
				m.textarea, tiCmd = m.textarea.Update(msg)
			} else {
				// Send message
				content := strings.TrimSpace(m.textarea.Value())
				if content != "" {
					m.messages = append(m.messages, Message{
						content: content,
						isUser:  true,
					})

					// Simple response
					m.messages = append(m.messages, Message{
						content: "Thanks for your message: " + content,
						isUser:  false,
					})

					// Update viewport with new messages
					m.viewport.SetContent(m.messageView())
					// Auto scroll to bottom
					m.viewport.GotoBottom()

					// Clear textarea
					m.textarea.Reset()
				}
				return m, nil
			}
		}
	}

	// Handle textarea updates
	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m Model) headerView() string {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Padding(0, 1).
		Width(m.viewport.Width).
		Render("Simple BubbleTea Chat")
}

func (m Model) footerView() string {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		Padding(1, 0).
		Width(m.viewport.Width).
		Render(m.textarea.View())
}

func (m Model) messageView() string {
	var sb strings.Builder

	for i, msg := range m.messages {
		if i > 0 {
			sb.WriteString("\n")
		}

		if msg.isUser {
			sb.WriteString(m.senderStyle.Render("You: "))
		} else {
			sb.WriteString(m.systemStyle.Render("System: "))
		}

		sb.WriteString(msg.content)
	}

	return sb.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
