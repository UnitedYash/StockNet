package friends

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
	"fmt"

)

// struct to hold a friend
type Friend struct {
	FriendEmail		string
	FriendName		string
}
// holds friends for update function
type FriendsLoadedMsg struct {
	friends []Friend
}
// holds error for update function
type FriendsLoadError struct {
    err error
}

// Model for the manage friend page
type ManageFriendsModel struct {
	BackPressed bool
	User        *auth.User
	Selected	int
	Friends		[]Friend
	Loading		bool
	Error		string

}

// returns initial stock list model
func NewManageFriendsPage(user *auth.User) *ManageFriendsModel {
	return &ManageFriendsModel{
		User: 		user,
		Friends:	[]Friend{},
		Selected:	0,
		Loading:	true,
	}
}
// returns initial command for the manage page to run
func (m *ManageFriendsModel) Init() tea.Cmd {
	return func() tea.Msg {
		// get all friends
		db := database.New().GetDB()
		
		rows, err := db.Query(`
			SELECT 
				a.email AS friend_email,
				a.name  AS friend_name
			FROM friendstatus f
			JOIN accounts a 
				ON a.email = (
					CASE 
						WHEN f.sender = $1 THEN f.receiver
						ELSE f.sender
					END
				)
			WHERE
				(f.sender = $1 OR f.receiver = $1)
				AND f.status = 'accepted';
		`, m.User.Email)

		if err != nil {
			return FriendsLoadError{err: err}
		}
		defer rows.Close()

		// store results within struct
		var friends []Friend
		for rows.Next() {
			var fr Friend
			if err := rows.Scan(&fr.FriendEmail, &fr.FriendName); err != nil {
				return FriendsLoadError{err: err} 
			}
			friends = append(friends, fr)

		}
		return FriendsLoadedMsg{friends: friends}
	}
}
// Handles incoming events and updates the model accordingly
func (m *ManageFriendsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FriendsLoadedMsg:
		m.Friends = msg.friends
		m.Loading = false
		m.Error = ""
	case FriendsLoadError:
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
				m.Selected = len(m.Friends) - 1
			}
		case "down", "j":
			if m.Selected < len(m.Friends) - 1 {
				m.Selected++
			} else {
				// at last option so wrap around to the top
				m.Selected = 0
			}
		case "r":
			//TODO: press r to remove a friend
			if len(m.Friends) == 0 {
				break
			}
			// get the highlighted friend 
			fri := m.Friends[m.Selected]
			db := database.New().GetDB()

			// change the relationship to status deleted
			_, err := db.Exec(`
				UPDATE friendstatus
				SET status = 'deleted',
					last_modified = NOW()
				WHERE 
					(
						(sender = $1 AND receiver = $2)
						OR
						(sender = $2 AND receiver = $1)
					)
					AND status = 'accepted'
			`, m.User.Email, fri.FriendEmail)
			if err != nil {
				m.Error = fmt.Sprintf("Error removing friend: %v", err)
				break
			}
			// remove friend from UI friends list
			m.Friends = append(m.Friends[:m.Selected], m.Friends[m.Selected+1:]...)

			// fix the indexing
			if m.Selected >= len(m.Friends) && len(m.Friends) > 0 {
				m.Selected = len(m.Friends) - 1
			}
			m.Error = ""


		}
	}
	return m, nil
}

func (m *ManageFriendsModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("🫂  Manage Friends") + "\n\n"

	if m.Loading {
		s += "Loading friends...\n"
	} else if m.Error != "" {
			s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if len(m.Friends) == 0 {
		s += "No Friends.\n"
	} else {
		for i, fri := range m.Friends {
			line := fmt.Sprintf("%s (%s)", fri.FriendEmail, fri.FriendName)
			if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+ line))
			} else {
				s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+ line))
			}
		}
	}

	s += styles.FooterStyle.Render("'r' to remove friends • ↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
