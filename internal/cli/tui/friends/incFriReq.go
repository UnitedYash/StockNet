package friends

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
	"fmt"
)

// struct to hold a incoming request
type IncFriendRequest struct {
	SenderEmail		string
	SenderName		string
}
// holds requests for update function
type IncRequestsLoadedMsg struct {
	requests []IncFriendRequest
}
// holds error for update function
type IncRequestsLoadError struct {
    err error
}

// Model for the incoming friend request page
type IncFriReqModel struct {
	BackPressed bool
	User        *auth.User
	Selected	int
	Requests	[]IncFriendRequest
	Loading		bool
	Error		string
}

// returns initial incoming friend request page model
func NewIncFriReqPage(user *auth.User) *IncFriReqModel {
	return &IncFriReqModel{
		User: 		user,
		Requests: 	[]IncFriendRequest{},
		Selected:	0,
		Loading:	true,
	}
}
// returns initial command for the incoming friend request page to run
func (m *IncFriReqModel) Init() tea.Cmd {
	// we get incoming requests from DB
	return func() tea.Msg {

		db := database.New().GetDB()
		// get senders email and their name
		rows, err := db.Query(`
			SELECT sender, name
			FROM friendstatus
			INNER JOIN accounts ON email = sender
			WHERE receiver = $1 AND status = 'pending'
		`, m.User.Email)

		if err != nil {
			return IncRequestsLoadError{err: err} 
		}
		defer rows.Close()

		// store within requests 
		var requests []IncFriendRequest
		for rows.Next() {
			var fr IncFriendRequest
			if err := rows.Scan(&fr.SenderEmail, &fr.SenderName); err != nil {
				return IncRequestsLoadError{err: err} 
			}
			requests = append(requests, fr)
		}
		return IncRequestsLoadedMsg{requests: requests}
	}
}
// Handles incoming events and updates the model accordingly
func (m *IncFriReqModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case IncRequestsLoadedMsg:
		m.Requests = msg.requests
		m.Loading = false
		m.Error = ""

	case IncRequestsLoadError:
		m.Loading = false
		m.Error = fmt.Sprintf("Error loading requests: %v", msg.err)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true
		case "up", "k":
			if m.Selected > 0 {
				// go up an option
				m.Selected--
			} else {
				// at the top so wrap around to bottom
				m.Selected = len(m.Requests) - 1
			}
		case "down", "j":
			if m.Selected < len(m.Requests) - 1 {
				m.Selected++
			} else {
				// at last option so wrap around to the top
				m.Selected = 0
			}
		case "a":
			// accept the friend request
			if len(m.Requests) == 0 {
        		break
    		}
			// get the highlighted request
			req := m.Requests[m.Selected]
    		db := database.New().GetDB()

			_, err := db.Exec(`
				UPDATE friendstatus
				SET status = 'accepted',
					last_modified = NOW()
				WHERE sender = $1 AND receiver = $2 AND status = 'pending'
			`, req.SenderEmail, m.User.Email)

			if err != nil {
				m.Error = fmt.Sprintf("Failed to reject request: %v", err)
				break
			}

			// remove from UI for that request
			m.Requests = append(m.Requests[:m.Selected], m.Requests[m.Selected+1:]...)
			// fix indexing of requests
			if m.Selected >= len(m.Requests) && m.Selected > 0 {
				m.Selected--
			}
			m.Error = ""

			
		case "r":
			// reject the friend request
			if len(m.Requests) == 0 {
        		break
    		}
			// get the highlighted request
			req := m.Requests[m.Selected]
    		db := database.New().GetDB()

			// mark the request as rejected and update last_modified so cooldown applies
			_, err := db.Exec(`
				UPDATE friendstatus
				SET status = 'rejected',
					last_modified = NOW()
				WHERE sender = $1 AND receiver = $2 AND status = 'pending'
			`, req.SenderEmail, m.User.Email)
			if err != nil {
				m.Error = fmt.Sprintf("Failed to reject request: %v", err)
				break
			}

			// Remove from UI list (it remains in DB with status='rejected')
			m.Requests = append(m.Requests[:m.Selected], m.Requests[m.Selected+1:]...)
			if m.Selected >= len(m.Requests) && m.Selected > 0 {
				m.Selected--
			}
			m.Error = ""


		}
	}
	return m, nil
}

func (m *IncFriReqModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("👋 Incoming Friend Requests") + "\n\n"

	if m.Loading {
		s += "Loading incoming requests...\n"
	} else if m.Error != "" {
			s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if len(m.Requests) == 0 {
		s += "No pending requests.\n"
	} else {
		for i, req := range m.Requests {
			line := fmt.Sprintf("%s (%s)", req.SenderName, req.SenderEmail)
			if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+ line))
			} else {
				s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+ line))
			}
		}
	}

	s += styles.FooterStyle.Render("'a' or 'r' to accept or reject • ↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}








