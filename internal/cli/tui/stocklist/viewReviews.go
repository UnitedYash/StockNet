package stocklist
import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/database"
	"StockNet/internal/auth"
	"fmt"
	"StockNet/internal/cli/tui/styles"
)


// model for viewing reviews of a stocklist
type ViewReviewsModel struct {
	StockList   StockList
    Selected    int
    BackPressed bool
    Error       string
	User        *auth.User
	Reviews      []Review   
	Loading		bool

}

type Review struct {
    UserID    	uint32
    Text      	string
	Deleteable	bool		// can the current user have permission to delete the review?
}
// struct to hold reivews during loading
type reviewsLoadedMsg struct{ reviews []Review }
type reviewsErrorMsg struct{ err error }

// returns initial view review  model
func NewViewReviewsPage(stockList StockList, user *auth.User) *ViewReviewsModel {
	return &ViewReviewsModel{
		Reviews:    	[]Review{},
		Selected:      	0,
		BackPressed:   	false,
		User: 			user,
		StockList:		stockList,
		Loading:		true,
	}
}

// returns initial command for to get all visible reviews
func (m *ViewReviewsModel) Init() tea.Cmd {
	return func() tea.Msg {
		reviews, err := GetReviewsForStockList(m.StockList, m.User)
		if err != nil {
			return reviewsErrorMsg{err: err}
		}
		return reviewsLoadedMsg{reviews: reviews}
	}
}


// handle incoming events and update the model accordinly
func (m *ViewReviewsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case reviewsLoadedMsg:
		m.Reviews = msg.reviews
		m.Error = ""
		m.Loading = false
	case reviewsErrorMsg:
		m.Error = fmt.Sprintf("Error loading reviews: %v", msg.err)
		m.Loading = false
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true
		case "up", "k":
			if m.Selected > 0 {
				m.Selected--
			} else {
				m.Selected = len(m.Reviews) - 1
			}
		case "down", "j":
			if m.Selected < len(m.Reviews)-1 {
				m.Selected++
			} else {
				m.Selected = 0
			}
		case "d": 
			// if the select review can be deleted then delete
			if len(m.Reviews) > 0 && m.Reviews[m.Selected].Deleteable {
				r := m.Reviews[m.Selected]
				err := DeleteReview(r.UserID, m.StockList.StockListID)
				if err != nil {
					m.Error = fmt.Sprintf("Failed to delete review: %v", err)
				} else {
					// remove in the gui aswell
					m.Reviews = append(m.Reviews[:m.Selected], m.Reviews[m.Selected+1:]...)
					if m.Selected >= len(m.Reviews) && len(m.Reviews) > 0 {
						m.Selected = len(m.Reviews) - 1
					}
				}
			}
		}
	}
	return m, nil
}

// renders the the gui based on the model
func (m *ViewReviewsModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📰 View Reviews for " + m.StockList.Name) + "\n\n"
	if m.Loading {
		s += "Loading in Reviews!\n"
	} else if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if len(m.Reviews) == 0 {
		s += "No reviews found. Womp Womp\n"
	} else {
		for i, r := range m.Reviews {
			line := fmt.Sprintf("User %d: %s", r.UserID, r.Text)
			if r.Deleteable {
				line += " [d to delete]"
			}
			if i == m.Selected {
				s += styles.SelectedStyle.Render(line) + "\n"
			} else {
				s += line + "\n"
			}
		}
	}
	s += "\n" + styles.FooterStyle.Render("'d' to delete review (if allowed) • ↑/↓ or k/j to navigate • 'Ctrl+b' or 'Esc' to go back") + "\n\n"
	return s

}

// get all reviews that the current user can see
func GetReviewsForStockList(stocklist StockList, currentUser *auth.User) ([]Review, error) {

	 // get the owner user_id of the stock
	 // this is needed to determine if they can see all reviews (if they are owner) or the ones they made only (non-owner)
    var ownerID uint32
	db := database.New().GetDB()

    err := db.QueryRow(`SELECT user_id FROM stocklist WHERE stocklist_id=$1`, stocklist.StockListID).Scan(&ownerID)
    if err != nil {
        return nil, err
    }
	// get all reivews 
	rows, err := db.Query(`
        SELECT user_id, text
        FROM reviews
        WHERE stocklist_id=$1
    `, stocklist.StockListID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

	reviews := []Review{}

	for rows.Next() {
        var r Review
        if err := rows.Scan(&r.UserID, &r.Text); err != nil {
            return nil, err
        }

        // determine if current user can see this review
		// if it private, not the owner of the stocklist and not the owner of the review, dont add
        if stocklist.Visibility == "private" && currentUser.UserID != ownerID && currentUser.UserID != r.UserID {
            continue 
        }

        // can delete if owner or reviewer
        r.Deleteable = currentUser.UserID == ownerID || currentUser.UserID == r.UserID
        reviews = append(reviews, r)
    }
    return reviews, nil
}

// delete a review
func DeleteReview(userID uint32, stocklistID int) error {
	db := database.New().GetDB()
	_, err := db.Exec(`DELETE FROM reviews WHERE user_id=$1 AND stocklist_id=$2`, userID, stocklistID)
	return err
}