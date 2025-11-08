package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type RegisterModel struct {
	backPressed bool
	step        int // 0: email, 1: password, 2: confirm password
	email       string
	password    string
	confirmPwd  string
	message     string
}

func NewRegister() *RegisterModel {
	return &RegisterModel{
		step: 0,
	}
}

func (m *RegisterModel) Init() tea.Cmd {
	return nil
}

func (m *RegisterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.step == 0 && m.email != "" {
				m.step = 1
			} else if m.step == 1 && m.password != "" {
				m.step = 2
			} else if m.step == 2 && m.confirmPwd != "" {
				m.message = "✓ Registration successful!"
			}
		case "backspace":
			if m.step == 0 && len(m.email) > 0 {
				m.email = m.email[:len(m.email)-1]
			} else if m.step == 1 && len(m.password) > 0 {
				m.password = m.password[:len(m.password)-1]
			} else if m.step == 2 && len(m.confirmPwd) > 0 {
				m.confirmPwd = m.confirmPwd[:len(m.confirmPwd)-1]
			}
		case "b", "esc":
			m.backPressed = true
		default:
			if len(msg.String()) == 1 {
				if m.step == 0 {
					m.email += msg.String()
				} else if m.step == 1 {
					m.password += msg.String()
				} else if m.step == 2 {
					m.confirmPwd += msg.String()
				}
			}
		}
	}
	return m, nil
}

func (m *RegisterModel) View() string {
	s := "\n"
	s += TitleStyle.Render("📝 Register") + "\n\n"

	if m.step == 0 {
		s += InputStyle.Render("Enter email: "+m.email) + "\n"
	} else if m.step == 1 {
		s += InputStyle.Render("Email: "+m.email) + "\n"
		s += InputStyle.Render("Enter password: "+hidePassword(m.password)) + "\n"
	} else {
		s += InputStyle.Render("Email: "+m.email) + "\n"
		s += InputStyle.Render("Password: "+hidePassword(m.password)) + "\n"
		s += InputStyle.Render("Confirm password: "+hidePassword(m.confirmPwd)) + "\n"
	}

	if m.message != "" {
		s += SuccessStyle.Render(m.message) + "\n"
	}

	s += FooterStyle.Render("Press Enter to continue • 'b' or 'Esc' to go back") + "\n\n"
	return s
}
