package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"github.com/charmbracelet/lipgloss"
)

// AppState represents different screens in the app
type AppState int

const (
	MainMenuState AppState = iota 	// 0
	LoginState						// 1
	RegisterState					// 2
	ConfigureState					// 3
	HomePageState					// 4
	PortfolioState					// 5
	StockListState					// 6
	StockAnalysisState				// 7
	SocialState						// 8
	CurrentStocksState				// 9
	SearchStockState				// 10
	StockDetailsState				// 11
	SendFriReqState					// 12
	ViewFriReqState					// 13
	IncFriReqState					// 14
	OutFriReqState					// 15
	ViewPortfoliosState				// 16
	CreatePortfolioState			// 17
)

// AppModel is the root model for the entire app
type AppModel struct {
	state        		AppState
	currentUser 		auth.User // currently logged-in user
	mainMenu    		*MainMenuModel
	login       		*LoginModel
	register    		*RegisterModel
	configure   		*ConfigureModel
	homepage    		*HomePageModel
	portfolio			*PortfolioModel
	viewPortfolios		*ViewPortfoliosModel
	createPortfolio		*CreatePortfolioModel
	stockList			*StockListModel
	stockAnalysis 		*StockAnalysisModel
	social				*SocialModel
	currentStocks		*CurrentStocksModel
	searchStock			*SearchStockModel
	stockDetails		*StockDetailsModel
	sendFriReq			*SendFriReqModel
	viewFriReq			*ViewFriReqModel
	incFriReq			*IncFriReqModel
	outFriReq 			*OutFriReqModel

}

// NewAppModel creates a new app model
func NewAppModel() *AppModel {
	return &AppModel{
		state:     			MainMenuState,
		mainMenu:  			NewMainMenu(),
		login:     			NewLogin(),
		register:  			NewRegister(),
		configure: 			NewConfigure(),
		homepage:  			NewHomePage(nil),
		portfolio: 			newPortfolioPage(),
		viewPortfolios:		newViewPortfoliosPageWithUserID(0), // Will be set with actual user ID
		createPortfolio:		newCreatePortfolioPageWithUserID(0), // Will be set with actual user ID
		stockList: 			newStockListPage(),
		stockAnalysis: 		newStockAnalysisPage(),
		social:				newSocialPage(nil),
		currentStocks:		newCurrentStocksPage(),
		searchStock:		newSearchStockPage(),
		stockDetails:		nil,
		sendFriReq:			newSendFriReqPage(nil),
		viewFriReq:			newViewFriReqPage(nil),
		incFriReq:			newIncFriReqPage(nil),
		outFriReq:			newOutFriReqPage(nil),

	}
}
// returns intial command for the application to run
func (m *AppModel) Init() tea.Cmd {
	// Note: For now, we don't have any initial I/O commands to do. nil = "no command"
	return nil
}
// handles incoming events and updates root model and runs corresponding commands
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit key
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Route to the appropriate state handler
	switch m.state {
	// give msg input to Update method of the currently active model
	// and changes app model state if a button/option (that is meant for switching views) gets triggered

	case MainMenuState:
		menu, cmd := m.mainMenu.Update(msg)
		m.mainMenu = menu.(*MainMenuModel)

		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.mainMenu.selected >= 0 && m.mainMenu.selected < len(m.mainMenu.options) {
			option := m.mainMenu.options[m.mainMenu.selected]
			if m.mainMenu.confirmed {
				m.mainMenu.confirmed = false
				switch option {
				case "Login":
					m.state = LoginState
					m.login = NewLogin()
				case "Register":
					m.state = RegisterState
					m.register = NewRegister()
				case "Configure":
					m.state = ConfigureState
					m.configure = NewConfigure()
				case "Quit":
					return m, tea.Quit
				}
			}
		}
		return m, cmd

	case LoginState:
		login, cmd := m.login.Update(msg)
		m.login = login.(*LoginModel)
		// Check if user successfully logged in
		if m.login.GetUser() != nil {
			m.currentUser = *m.login.GetUser()
			m.state = HomePageState
			m.homepage = NewHomePage(m.login.GetUser()) // reset homepage
			m.login = NewLogin() // reset login form
		}
		// Go back to main view from Login view
		if m.login.backPressed {
			m.login.backPressed = false
			m.state = MainMenuState
		}
		return m, cmd

	case RegisterState:
		register, cmd := m.register.Update(msg)
		m.register = register.(*RegisterModel)
		// Check if user successfully registered
		if m.register.GetUser() != nil {
			m.currentUser = *m.register.GetUser()
			m.state = HomePageState
			m.homepage = NewHomePage(m.register.GetUser()) // reset homepage
			m.register = NewRegister() // reset register form
		}
		// Go back to main view from register view
		if m.register.backPressed {
			m.register.backPressed = false
			m.state = MainMenuState
		}
		return m, cmd

	case ConfigureState:
		configure, cmd := m.configure.Update(msg)
		m.configure = configure.(*ConfigureModel)

		// Go back to home view from configure view
		if m.configure.backPressed {
			m.configure.backPressed = false
			m.state = MainMenuState
		}
		return m, cmd

	case HomePageState:
		homepage, cmd := m.homepage.Update(msg)
		m.homepage = homepage.(*HomePageModel)

		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.homepage.selected >= 0 && m.homepage.selected < len(m.homepage.options) {
			option := m.homepage.options[m.homepage.selected]
			if m.homepage.confirmed {
				m.homepage.confirmed = false
				switch option {
				case "My Portfolios":
					m.state = PortfolioState
					m.portfolio = newPortfolioPage()
				case "My Stock Lists":
					m.state = StockListState
					m.stockList = newStockListPage()
				case "Stock Data & Analysis":
					m.state = StockAnalysisState
					m.stockAnalysis = newStockAnalysisPage()
				case "Friends & Social":
					m.state = SocialState
					m.social = newSocialPage(m.homepage.GetUser())
				}
			}
		}

		// Go back to main menu (logout) from HomePage
		if m.homepage.backPressed {
			m.homepage.backPressed = false
			m.state = MainMenuState
			m.currentUser = auth.User{} // clear current user
		}
		return m, cmd
	case PortfolioState:
		portfolio, cmd := m.portfolio.Update(msg)
		m.portfolio = portfolio.(*PortfolioModel)

		// Check if user pressed Enter to select an option
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			option := m.portfolio.GetSelectedOption()
			userID := int(m.currentUser.UserID)
			if option == 0 {
				// View Portfolios selected
				m.state = ViewPortfoliosState
				m.viewPortfolios = newViewPortfoliosPageWithUserID(userID)
				cmd = m.viewPortfolios.Init()
			} else if option == 1 {
				// Create Portfolio selected
				m.state = CreatePortfolioState
				m.createPortfolio = newCreatePortfolioPageWithUserID(userID)
			}
		}

		// Go back to homepage from portfolio page
		if m.portfolio.backPressed {
			m.portfolio.backPressed = false
			m.state = HomePageState
		}
		return m, cmd
	case ViewPortfoliosState:
		viewPortfolios, cmd := m.viewPortfolios.Update(msg)
		m.viewPortfolios = viewPortfolios.(*ViewPortfoliosModel)
		// Go back to portfolio page from view portfolios page
		if m.viewPortfolios.backPressed {
			m.viewPortfolios.backPressed = false
			m.state = PortfolioState
			m.portfolio = newPortfolioPage()
		}
		return m, cmd
	case CreatePortfolioState:
		createPortfolio, cmd := m.createPortfolio.Update(msg)
		m.createPortfolio = createPortfolio.(*CreatePortfolioModel)
		// Go back to portfolio page from create portfolio page
		if m.createPortfolio.backPressed {
			m.createPortfolio.backPressed = false
			m.state = PortfolioState
			m.portfolio = newPortfolioPage()
		}
		return m, cmd
	case StockListState:
		stockList, cmd := m.stockList.Update(msg)
		m.stockList = stockList.(*StockListModel)
		// Go back to homepage from stock list page
		if m.stockList.backPressed {
			m.stockList.backPressed = false
			m.state = HomePageState
		}
		return m, cmd
	case StockAnalysisState:
		stockAnalysis, cmd := m.stockAnalysis.Update(msg)
		m.stockAnalysis = stockAnalysis.(*StockAnalysisModel)

		// Handle menu option selection
		if m.stockAnalysis.selected >= 0 && m.stockAnalysis.selected < len(m.stockAnalysis.options) {
			option := m.stockAnalysis.options[m.stockAnalysis.selected]
			if m.stockAnalysis.confirmed {
				m.stockAnalysis.confirmed = false
				switch option {
				case "View Current Stocks":
					m.state = CurrentStocksState
					m.currentStocks = newCurrentStocksPage()
					// Return Init command to fetch stocks data
					cmd = m.currentStocks.Init()
				case "Search Stock":
					m.state = SearchStockState
					m.searchStock = newSearchStockPage()
				}
			}
		}

		// Go back to homepage from Stock Data & Analysis page
		if m.stockAnalysis.backPressed {
			m.stockAnalysis.backPressed = false
			m.state = HomePageState
		}
		return m, cmd
	case SocialState:
		social, cmd := m.social.Update(msg)
		m.social = social.(*SocialModel)

		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.social.selected >= 0 && m.social.selected < len(m.social.options) {
			option := m.social.options[m.social.selected]
			if m.social.confirmed {
				m.social.confirmed = false
				switch option {
				case "Send Friend Request":
					m.state = SendFriReqState
					m.sendFriReq = newSendFriReqPage(m.social.GetUser())
				case "View Friends Requests":
					m.state = ViewFriReqState
					m.viewFriReq = newViewFriReqPage(m.social.GetUser())
				}
			}
		}
		// Go back to homepage from social page
		if m.social.backPressed {
			m.social.backPressed = false
			m.state = HomePageState
		}
		return m, cmd

	case CurrentStocksState:
		currentStocks, cmd := m.currentStocks.Update(msg)
		m.currentStocks = currentStocks.(*CurrentStocksModel)
		// Go back to stock analysis page from current stocks page
		if m.currentStocks.backPressed {
			m.currentStocks.backPressed = false
			m.state = StockAnalysisState
			m.stockAnalysis = newStockAnalysisPage()
		}
		return m, cmd
	case SearchStockState:
		searchStock, cmd := m.searchStock.Update(msg)
		m.searchStock = searchStock.(*SearchStockModel)
		// Check if user confirmed search
		if m.searchStock.confirmed {
			m.searchStock.confirmed = false
			symbol := m.searchStock.GetSymbol()
			m.state = StockDetailsState
			m.stockDetails = newStockDetailsPage(symbol)
			// Return Init command to fetch historical data
			cmd = m.stockDetails.Init()
		}
		// Go back to stock analysis page
		if m.searchStock.backPressed {
			m.searchStock.backPressed = false
			m.state = StockAnalysisState
			m.searchStock = newSearchStockPage()
		}
		return m, cmd
	case StockDetailsState:
		stockDetails, cmd := m.stockDetails.Update(msg)
		m.stockDetails = stockDetails.(*StockDetailsModel)
		// Go back to search stock page
		if m.stockDetails.backPressed {
			m.stockDetails.backPressed = false
			m.state = SearchStockState
			m.searchStock = newSearchStockPage()
		}
		return m, cmd
	case SendFriReqState:
		sendFriReq, cmd := m.sendFriReq.Update(msg)
		m.sendFriReq = sendFriReq.(*SendFriReqModel)
		// Go back to friend and social from send friend request page
		if m.sendFriReq.backPressed {
			m.sendFriReq.backPressed = false
			m.state = SocialState
		}
		return m, cmd
	case ViewFriReqState:
		viewFriReq, cmd := m.viewFriReq.Update(msg)
		m.viewFriReq = viewFriReq.(*ViewFriReqModel)
		// Go back to friend and social from view friend request page
		if m.viewFriReq.backPressed {
			m.viewFriReq.backPressed = false
			m.state = SocialState
		}
		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.viewFriReq.selected >= 0 && m.viewFriReq.selected < len(m.viewFriReq.options) {
			option := m.viewFriReq.options[m.viewFriReq.selected]
			if m.viewFriReq.confirmed {
				m.viewFriReq.confirmed = false
				switch option {
				case "Incoming Requests (Accept / Reject)":
					m.state = IncFriReqState
					m.incFriReq = newIncFriReqPage(m.viewFriReq.GetUser())
					cmd = m.incFriReq.Init()
				case "Outgoing Requests (Cancel)":
					m.state = OutFriReqState
					m.outFriReq = newOutFriReqPage(m.viewFriReq.GetUser())
					cmd = m.outFriReq.Init()

				}
			}
		}
		return m, cmd
	case IncFriReqState:
		incFriReq, cmd := m.incFriReq.Update(msg)
		m.incFriReq = incFriReq.(*IncFriReqModel)
		// Go back to view friend request page from incoming view page
		if m.incFriReq.backPressed {
			m.incFriReq.backPressed = false
			m.state = ViewFriReqState
		}
		return m, cmd
	case OutFriReqState:
		outFriReq, cmd := m.outFriReq.Update(msg)
		m.outFriReq = outFriReq.(*OutFriReqModel)
		// Go back to view friend request page from outgoing view page
		if m.outFriReq.backPressed {
			m.outFriReq.backPressed = false
			m.state = ViewFriReqState
		}
		return m, cmd
	}
	return m, nil
}

// Render the current model view
func (m *AppModel) View() string {
	switch m.state {
	case MainMenuState:
		return m.mainMenu.View()
	case LoginState:
		return m.login.View()
	case RegisterState:
		return m.register.View()
	case ConfigureState:
		return m.configure.View()
	case HomePageState:
		return m.homepage.View()
	case PortfolioState:
		return m.portfolio.View()
	case ViewPortfoliosState:
		return m.viewPortfolios.View()
	case CreatePortfolioState:
		return m.createPortfolio.View()
	case StockListState:
		return m.stockList.View()
	case StockAnalysisState:
		return m.stockAnalysis.View()
	case SocialState:
		return m.social.View()
	case CurrentStocksState:
		return m.currentStocks.View()
	case SearchStockState:
		return m.searchStock.View()
	case StockDetailsState:
		return m.stockDetails.View()
	case SendFriReqState:
		return m.sendFriReq.View()
	case ViewFriReqState:
		return m.viewFriReq.View()
	case IncFriReqState:
		return m.incFriReq.View()
	case OutFriReqState:
		return m.outFriReq.View()
	}
	return ""
}

// StartApp starts the interactive CLI app
func StartApp() error {
	p := tea.NewProgram(NewAppModel())
	_, err := p.Run()
	return err
}

// Common styles
var (
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("3")).
		PaddingLeft(2).
		PaddingRight(2)

	SelectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("2")).
		Background(lipgloss.Color("8")).
		PaddingLeft(2)

	UnselectedStyle = lipgloss.NewStyle().
		PaddingLeft(2)

	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		PaddingLeft(2)

	FooterStyle = lipgloss.NewStyle().
		Faint(true).
		PaddingTop(1).
		PaddingLeft(2)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		PaddingLeft(2)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		PaddingLeft(2)

	InputStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("6"))
)
