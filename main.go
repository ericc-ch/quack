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
	messages          []Message
	viewport          viewport.Model
	textarea          textarea.Model
	ready             bool
	senderStyle       lipgloss.Style
	systemStyle       lipgloss.Style
	focusState        string // "textarea", "viewport", or "message"
	selectedMessageID int    // Track which message is selected
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
		messages:          []Message{},
		textarea:          ta,
		viewport:          viewport.New(0, 0),
		senderStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		systemStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		focusState:        "textarea", // Start with textarea focused
		selectedMessageID: -1,         // No message selected initially
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
		case tea.KeyTab:
			// Cycle through focus states: textarea -> viewport -> message -> textarea
			switch m.focusState {
			case "textarea":
				m.focusState = "viewport"
				m.textarea.Blur()
				m.selectedMessageID = -1
			case "viewport":
				if len(m.messages) > 0 {
					m.focusState = "message"
					m.selectedMessageID = 0
				} else {
					m.focusState = "textarea"
					m.textarea.Focus()
				}
			case "message":
				m.focusState = "textarea"
				m.textarea.Focus()
				m.selectedMessageID = -1
			}
			return m, nil
		case tea.KeyEnter:
			if m.focusState == "textarea" {
				if key.Alt {
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
		// Navigation between messages when in message focus mode
		case tea.KeyUp:
			if m.focusState == "message" && m.selectedMessageID > 0 {
				m.selectedMessageID--
				m.viewport.SetContent(m.messageView())
				return m, nil
			}
		case tea.KeyDown:
			if m.focusState == "message" && m.selectedMessageID < len(m.messages)-1 {
				m.selectedMessageID++
				m.viewport.SetContent(m.messageView())
				return m, nil
			}
		}
		
		// Only pass key events to the currently focused component
		if m.focusState == "textarea" {
			m.textarea, tiCmd = m.textarea.Update(msg)
		} else if m.focusState == "viewport" {
			m.viewport, vpCmd = m.viewport.Update(msg)
		}
		// When in message focus mode, we've already handled navigation above
		
		return m, tea.Batch(tiCmd, vpCmd)
	}

	// For non-key messages, update both components
	if _, ok := msg.(tea.KeyMsg); !ok {
		m.textarea, tiCmd = m.textarea.Update(msg)
		m.viewport, vpCmd = m.viewport.Update(msg)
		return m, tea.Batch(tiCmd, vpCmd)
	}
	
	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m Model) headerView() string {
	title := "Simple BubbleTea Chat"
	var help string
	
	switch m.focusState {
	case "textarea":
		help = "(Tab to switch focus) [Input Mode]"
	case "viewport":
		help = "(Tab to switch focus) [Scroll Mode ↑/↓]"
	case "message":
		help = "(Tab to switch focus) [Message Selection Mode ↑/↓]"
	}
	
	titleStyle := lipgloss.NewStyle().Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	
	header := lipgloss.JoinHorizontal(
		lipgloss.Center,
		titleStyle.Render(title),
		" ",
		helpStyle.Render(help),
	)
	
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Padding(0, 1).
		Width(m.viewport.Width).
		Render(header)
}

func (m Model) footerView() string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		Padding(1, 0).
		Width(m.viewport.Width)
	
	// Add visual indicator of focus
	if m.focusState == "textarea" {
		style = style.BorderForeground(lipgloss.Color("6")) // Highlight border when focused
	}
	
	return style.Render(m.textarea.View())
}

func (m Model) messageView() string {
	var sb strings.Builder

	// Add visual indicator of focus state
	if m.focusState == "viewport" && m.ready {
		sb.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Render("Messages (scrollable) ↑/↓\n\n"))
	} else if m.focusState == "message" && m.ready {
		sb.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).
			Bold(true).
			Render("Message Selection Mode (↑/↓ to navigate)\n\n"))
	}

	for i, msg := range m.messages {
		if i > 0 {
			sb.WriteString("\n")
		}

		// Style for sender prefix
		senderPrefix := ""
		if msg.isUser {
			senderPrefix = "You: "
		} else {
			senderPrefix = "System: "
		}

		// Determine if this message is the selected one
		if m.focusState == "message" && i == m.selectedMessageID {
			// Highlight the selected message
			selectedStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("4")).
				Foreground(lipgloss.Color("15")).
				Bold(true)
			
			// Combine the prefix and content with highlighting
			sb.WriteString(selectedStyle.Render(senderPrefix + msg.content))
		} else {
			// Regular styling for non-selected messages
			if msg.isUser {
				sb.WriteString(m.senderStyle.Render(senderPrefix))
			} else {
				sb.WriteString(m.systemStyle.Render(senderPrefix))
			}
			sb.WriteString(msg.content)
		}
	}

	return sb.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
