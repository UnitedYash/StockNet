package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
)

type RegisterModel struct {
	backPressed bool
	step        int // 0: name, 1: email, 2: password, 3: confirm password
	name        string
	email       string
	password    string
	confirmPwd  string
	message     string
	user        *auth.User
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
			if m.step == 0 && m.name != "" {
				m.step = 1
			} else if m.step == 1 && m.email != "" {
				m.step = 2
			} else if m.step == 2 && m.password != "" {
				m.step = 3
			} else if m.step == 3 && m.confirmPwd != "" {
				if m.password != m.confirmPwd {
					m.message = "✗ Passwords do not match!"
				} else {
					user, err := auth.Register(m.email, m.password, m.name)
					if err != nil {
						m.message = "✗ " + err.Error()
					} else {
						m.user = user
						m.message = "✓ Registration successful!"
					}
				}
			}
		case "backspace":
			if m.step == 0 && len(m.name) > 0 {
				m.name = m.name[:len(m.name)-1]
			} else if m.step == 1 && len(m.email) > 0 {
				m.email = m.email[:len(m.email)-1]
			} else if m.step == 2 && len(m.password) > 0 {
				m.password = m.password[:len(m.password)-1]
			} else if m.step == 3 && len(m.confirmPwd) > 0 {
				m.confirmPwd = m.confirmPwd[:len(m.confirmPwd)-1]
			}
		case "b", "esc":
			m.backPressed = true
		default:
			if len(msg.String()) == 1 {
				if m.step == 0 {
					m.name += msg.String()
				} else if m.step == 1 {
					m.email += msg.String()
				} else if m.step == 2 {
					m.password += msg.String()
				} else if m.step == 3 {
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
		s += InputStyle.Render("Enter name: "+m.name) + "\n"
	} else if m.step == 1 {
		s += InputStyle.Render("Name: "+m.name) + "\n"
		s += InputStyle.Render("Enter email: "+m.email) + "\n"
	} else if m.step == 2 {
		s += InputStyle.Render("Name: "+m.name) + "\n"
		s += InputStyle.Render("Email: "+m.email) + "\n"
		s += InputStyle.Render("Enter password: "+hidePassword(m.password)) + "\n"
	} else {
		s += InputStyle.Render("Name: "+m.name) + "\n"
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

func (m *RegisterModel) GetUser() *auth.User {
	return m.user
}
