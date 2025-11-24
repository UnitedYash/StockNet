package shared

import (
	"StockNet/internal/config"
	"StockNet/internal/cli/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
)
// MOdel for configure page
type ConfigureModel struct {
	BackPressed bool
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
			m.BackPressed = true
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
	s += styles.TitleStyle.Render("⚙️  Configure VM IP") + "\n\n"

	s += styles.InputStyle.Render("Enter VM External IP: "+m.vmIP) + "\n"

	if m.message != "" {
		if m.saved {
			s += styles.SuccessStyle.Render(m.message) + "\n"
		} else {
			s += styles.ErrorStyle.Render(m.message) + "\n"
		}
	}

	s += styles.FooterStyle.Render("Press Enter to save • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
