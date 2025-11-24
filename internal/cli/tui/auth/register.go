package auth

import (
	tea "github.com/charmbracelet/bubbletea"
	userauth "StockNet/internal/auth"
	"StockNet/internal/cli/tui/styles"
)
// Model for register view
type RegisterModel struct {
	BackPressed bool
	step        int // 0: name, 1: email, 2: password, 3: confirm password
	name        string
	email       string
	password    string
	confirmPwd  string
	message     string
	User        *userauth.User
}

// returns intital register model 
func NewRegister() *RegisterModel {
	return &RegisterModel{
		step: 0,
	}
}
// returns initial command for the register page to run (nothing)
func (m *RegisterModel) Init() tea.Cmd {
	return nil
}

// Handles incoming events and updates the model accordingly
func (m *RegisterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// move to the next text field or register the user
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
					user, err := userauth.Register(m.email, m.password, m.name)
					if err != nil {
						m.message = "✗ " + err.Error()
					} else {
						m.User = user
						m.message = "✓ Registration successful!"
					}
				}
			}
		case "backspace":
			// remove last character typed in current text field
			if m.step == 0 && len(m.name) > 0 {
				m.name = m.name[:len(m.name)-1]
			} else if m.step == 1 && len(m.email) > 0 {
				m.email = m.email[:len(m.email)-1]
			} else if m.step == 2 && len(m.password) > 0 {
				m.password = m.password[:len(m.password)-1]
			} else if m.step == 3 && len(m.confirmPwd) > 0 {
				m.confirmPwd = m.confirmPwd[:len(m.confirmPwd)-1]
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		default:
			// any other key would indicate user typing in text field
			// also block any non-printable letters
			if len(msg.String()) == 1 {
				r := rune(msg.String()[0])
				if r >= 32 && r != 127 { // valid ASCII except for DEL
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
	}
	return m, nil
}
// renders UI based on current register model
func (m *RegisterModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📝 Register") + "\n\n"

	if m.step == 0 {
		s += styles.InputStyle.Render("Enter name: "+m.name) + "\n"
	} else if m.step == 1 {
		s += styles.InputStyle.Render("Name: "+m.name) + "\n"
		s += styles.InputStyle.Render("Enter email: "+m.email) + "\n"
	} else if m.step == 2 {
		s += styles.InputStyle.Render("Name: "+m.name) + "\n"
		s += styles.InputStyle.Render("Email: "+m.email) + "\n"
		s += styles.InputStyle.Render("Enter password: "+hidePassword(m.password)) + "\n"
	} else {
		s += styles.InputStyle.Render("Name: "+m.name) + "\n"
		s += styles.InputStyle.Render("Email: "+m.email) + "\n"
		s += styles.InputStyle.Render("Password: "+hidePassword(m.password)) + "\n"
		s += styles.InputStyle.Render("Confirm password: "+hidePassword(m.confirmPwd)) + "\n"
	}

	if m.message != "" {
		s += styles.SuccessStyle.Render(m.message) + "\n"
	}

	s += styles.FooterStyle.Render("Press Enter to continue • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
// returns registered (logged in) user
func (m *RegisterModel) GetUser() *userauth.User {
	return m.User
}
