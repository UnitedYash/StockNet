package tui

import (
	"StockNet/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)
// MOdel for configure page
type ConfigureModel struct {
	backPressed bool
	vmIP        string
	message     string
	saved       bool
}
// returns Configure initial modei
func NewConfigure() *ConfigureModel {
	// Load existing VM IP if available
	ip, _ := config.GetVMIP()
	return &ConfigureModel{
		vmIP: ip,
	}
}
// returns initial command for the configure page (No commands yet)
func (m *ConfigureModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates model accordingly
func (m *ConfigureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// saving VM IP
			if m.vmIP != "" {
				if err := config.SetVMIP(m.vmIP); err != nil {
					m.message = "✗ Error saving config"
				} else {
					m.message = "✓ VM IP saved successfully!"
					m.saved = true
				}
			}
		case "backspace":
			// removing last char from text field
			if len(m.vmIP) > 0 {
				m.vmIP = m.vmIP[:len(m.vmIP)-1]
			}
		case "ctrl+b", "esc":
			m.backPressed = true
		default:
			// any other key implies user typing in text field
			// also block any non-printable letters

			if len(msg.String()) == 1 {
				r := rune(msg.String()[0])
				if r >=32 && r != 127 {
					m.vmIP += msg.String()
				}

			}
		}
	}
	return m, nil
}
// renders UI based on the current version of the configure model
func (m *ConfigureModel) View() string {
	s := "\n"
	s += TitleStyle.Render("⚙️  Configure VM IP") + "\n\n"

	s += InputStyle.Render("Enter VM External IP: "+m.vmIP) + "\n"

	if m.message != "" {
		if m.saved {
			s += SuccessStyle.Render(m.message) + "\n"
		} else {
			s += ErrorStyle.Render(m.message) + "\n"
		}
	}

	s += FooterStyle.Render("Press Enter to save • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
