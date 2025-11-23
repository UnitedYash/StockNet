package friends

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
	"time"
	"fmt"


)

// Model for the send friend request page
type SendFriReqModel struct {
	BackPressed bool
	User        *auth.User
	email		string
	message		string

}

// returns initial send friend request model
func NewSendFriReqPage(user *auth.User) *SendFriReqModel {
	return &SendFriReqModel{
		User: user,
	}
}
// returns initial command for the send friend request page to run (nothing)
func (m *SendFriReqModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *SendFriReqModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true
		case "enter":
			if m.email != "" {
				// attempt to send a request

				// skip if they put in their own email
				if m.email == m.User.Email {
					m.message = "✗ Cannot send friend request to yourself"
					break
				}

				// get db instance
				db := database.New().GetDB()

				// get if recipient even exist
				var recipientUsername string
				err := db.QueryRow("SELECT name FROM accounts WHERE email=$1", m.email).Scan(&recipientUsername)
				if err != nil || recipientUsername == "" {
    				m.message = "✗ No user found with that email"
    				break
				}
				
				// get latest status and when it happend 
				var status string
				var lastModified time.Time

				// for determining who is the sender and receiver of the relationship
				// these are default values, changes depending if the current user is receiver or if sender
				var sender = m.User.Email
				var receiver = m.email

				err = db.QueryRow(
					"SELECT status, last_modified FROM friendstatus WHERE sender=$1 AND receiver=$2",
					m.User.Email, m.email,
				).Scan(&status, &lastModified)
				

				if err != nil {
					// skip error check if 0 rows returned
					if err.Error() != "sql: no rows in result set" {
						m.message = "✗ Database error: " + err.Error()
						break
					} else {
						// then no row exist, theres case they have a relationship but user is the receiver
						err = db.QueryRow(
							"SELECT status, last_modified FROM friendstatus WHERE sender=$2 AND receiver=$1",
							m.User.Email, m.email,
						).Scan(&status, &lastModified)
						if err != nil {
							if err.Error() != "sql: no rows in result set" {
								m.message = "✗ Database error: " + err.Error()
								break
							}
						}
						// row exist, swap sender and receiver
						sender = m.email
						receiver = m.User.Email
					} 
				}
				if err == nil { // row exists
					now := time.Now()
					cooldown := 5 * time.Minute
					if status == "pending" {
						m.message = "✗ Friend request already pending"
						break
					} else if (status == "rejected" || status == "deleted") && now.Sub(lastModified) < cooldown {
						m.message = "✗ Cannot re-send yet. Try again later"
						break
					} else if status == "accepted" {
						m.message = "💖 You're already friends!"
						break
					}

					// otherwise, can update the row the timer is past 5 minutes depending on 
					if sender == m.User.Email {
						// no swap happend, current user is the sender
						_, err = db.Exec(
							"UPDATE friendstatus SET status='pending', last_modified=NOW() WHERE sender=$1 AND receiver=$2",
							sender, receiver,
						)
					} else {
						// the current user is receiver and sending a request delete old one 
						_, err = db.Exec(
							"DELETE FROM friendstatus WHERE sender = $1 AND receiver = $2",
							sender, receiver,
						)
						// now insert a fresh one
						_, err = db.Exec(
							"INSERT INTO friendstatus(sender, receiver, status, last_modified) VALUES($1, $2, 'pending', NOW())",
							receiver, sender,
						)
					}

					if err != nil {
						m.message = "✗ Database error: " + err.Error()
					} else {
						m.message = fmt.Sprintf("✓ Friend request sent to %s", recipientUsername)
						m.email = "" // clear input
					}
				} else { // no row exists, insert new
						_, err = db.Exec(
							"INSERT INTO friendstatus(sender, receiver, status, last_modified) VALUES($1, $2, 'pending', NOW())",
							m.User.Email, m.email,
						)
				}
				if err != nil {
					m.message = "✗ Database error: " + err.Error()
				} else {
					m.message = fmt.Sprintf("✓ Friend request sent to %s", recipientUsername)
					m.email = "" // clear input
				}
				
			}
		case "backspace":
			if len(m.email) > 0 {
				m.email = m.email[:len(m.email)-1]
			} 
		default:
			// any other key would indicate user typing in the text field
			// also block any non-printable letters
			if len(msg.String()) == 1 {
				r := rune(msg.String()[0])
				if r >= 32 && r != 127 {
					m.email += msg.String()
				}
			}
		}
		
	}
	return m, nil
}

func (m *SendFriReqModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📨 Send Friend Request") + "\n\n"
	s += styles.InputStyle.Render("Enter email: "+m.email) + "\n"

	if m.message != "" {
		s += styles.SuccessStyle.Render(m.message) + "\n"
	}

	s += styles.FooterStyle.Render("'Ctrl + b' or 'Esc' to go back") + "\n\n"


	return s
}
