package tui

import (
	"StockNet/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

type ConfigureModel struct {
	backPressed bool
	vmIP        string
	message     string
	saved       bool
}

func NewConfigure() *ConfigureModel {
	// Load existing VM IP if available
	ip, _ := config.GetVMIP()
	return &ConfigureModel{
		vmIP: ip,
	}
}

func (m *ConfigureModel) Init() tea.Cmd {
	return nil
}

func (m *ConfigureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.vmIP != "" {
				if err := config.SetVMIP(m.vmIP); err != nil {
					m.message = "✗ Error saving config"
				} else {
					m.message = "✓ VM IP saved successfully!"
					m.saved = true
				}
			}
		case "backspace":
			if len(m.vmIP) > 0 {
				m.vmIP = m.vmIP[:len(m.vmIP)-1]
			}
		case "b", "esc":
			m.backPressed = true
		default:
			if len(msg.String()) == 1 {
				m.vmIP += msg.String()
			}
		}
	}
	return m, nil
}

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

	s += FooterStyle.Render("Press Enter to save • 'b' or 'Esc' to go back") + "\n\n"
	return s
}
