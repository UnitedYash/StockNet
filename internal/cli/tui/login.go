package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
)

type LoginModel struct {
	backPressed bool
	step        int // 0: email, 1: password
	email       string
	password    string
	message     string
	user        *auth.User
}

func NewLogin() *LoginModel {
	return &LoginModel{
		step: 0,
	}
}

func (m *LoginModel) Init() tea.Cmd {
	return nil
}

func (m *LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.step == 0 && m.email != "" {
				m.step = 1
			} else if m.step == 1 && m.password != "" {
				user, err := auth.Login(m.email, m.password)
				if err != nil {
					m.message = "✗ " + err.Error()
				} else {
					m.user = user
					m.message = "✓ Login successful!"
				}
			}
		case "backspace":
			if m.step == 0 && len(m.email) > 0 {
				m.email = m.email[:len(m.email)-1]
			} else if m.step == 1 && len(m.password) > 0 {
				m.password = m.password[:len(m.password)-1]
			}
		case "ctrl+b", "esc":
			m.backPressed = true
		default:
			if len(msg.String()) == 1 {
				if m.step == 0 {
					m.email += msg.String()
				} else if m.step == 1 {
					m.password += msg.String()
				}
			}
		}
	}
	return m, nil
}

func (m *LoginModel) View() string {
	s := "\n"
	s += TitleStyle.Render("🔐 Login") + "\n\n"

	if m.step == 0 {
		s += InputStyle.Render("Enter email: "+m.email) + "\n"
	} else {
		s += InputStyle.Render("Email: "+m.email) + "\n"
		s += InputStyle.Render("Enter password: "+hidePassword(m.password)) + "\n"
	}

	if m.message != "" {
		s += SuccessStyle.Render(m.message) + "\n"
	}

	s += FooterStyle.Render("Press Enter to continue • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

func hidePassword(password string) string {
	result := ""
	for range password {
		result += "•"
	}
	return result
}

func (m *LoginModel) GetUser() *auth.User {
	return m.user
}
