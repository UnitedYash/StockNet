package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	tuiauth "StockNet/internal/cli/tui/auth"
	"StockNet/internal/cli/tui/friends"
	"StockNet/internal/cli/tui/portfolio"
	"StockNet/internal/cli/tui/shared"
	"StockNet/internal/cli/tui/stock"
	"StockNet/internal/cli/tui/stocklist"
	"StockNet/internal/database"
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
	ViewSpecPortfolioState			// 18
	ViewHoldingsState				// 19
	BuyStockSearchState				// 20
	BuyStockState					// 21
	ManageFriendsState				// 22
	CreateStockListState			// 23
	ViewStockListsState				// 24
)

// AppModel is the root model for the entire app
type AppModel struct {
	state        		AppState
	currentUser 		auth.User // currently logged-in user
	mainMenu    		*shared.MainMenuModel
	login       		*tuiauth.LoginModel
	register    		*tuiauth.RegisterModel
	configure   		*shared.ConfigureModel
	homepage    		*shared.HomePageModel
	portfolio			*portfolio.PortfolioModel
	viewPortfolios		*portfolio.ViewPortfoliosModel
	createPortfolio		*portfolio.CreatePortfolioModel
	stockList			*stocklist.StockListModel
	stockAnalysis 		*stock.StockAnalysisModel
	social				*shared.SocialModel
	currentStocks		*stock.CurrentStocksModel
	searchStock			*stock.SearchStockModel
	stockDetails		*stock.StockDetailsModel
	sendFriReq			*friends.SendFriReqModel
	viewFriReq			*friends.ViewFriReqModel
	incFriReq			*friends.IncFriReqModel
	outFriReq 			*friends.OutFriReqModel
	viewSpecPortfolio	*portfolio.ViewSpecPortfolioModel
	viewHoldings		*stock.ViewHoldingsModel
	buyStockSearch		*stock.BuyStockSearchModel
	buyStock			*stock.BuyStockModel
	manageFriends		*friends.ManageFriendsModel
	createStockList		*stocklist.CreateStockListModel
	viewMyStockList		*stocklist.ViewStockListsModel
}

// NewAppModel creates a new app model
func NewAppModel() *AppModel {
	return &AppModel{
		state:     			MainMenuState,
		mainMenu:  			shared.NewMainMenu(),
		login:     			tuiauth.NewLogin(),
		register:  			tuiauth.NewRegister(),
		configure: 			shared.NewConfigure(),
		homepage:  			shared.NewHomePage(nil),
		portfolio: 			portfolio.NewPortfolioPage(),
		viewPortfolios:		portfolio.NewViewPortfoliosPageWithUserID(0), // Will be set with actual user ID
		createPortfolio:		portfolio.NewCreatePortfolioPageWithUserID(0), // Will be set with actual user ID
		stockList: 			stocklist.NewStockListPage(nil),
		stockAnalysis: 		stock.NewStockAnalysisPage(),
		social:				shared.NewSocialPage(nil),
		currentStocks:		stock.NewCurrentStocksPage(),
		searchStock:		stock.NewSearchStockPage(),
		stockDetails:		nil,
		sendFriReq:			friends.NewSendFriReqPage(nil),
		viewFriReq:			friends.NewViewFriReqPage(nil),
		incFriReq:			friends.NewIncFriReqPage(nil),
		outFriReq:			friends.NewOutFriReqPage(nil),
		viewSpecPortfolio:	nil,
		viewHoldings:		nil,
		buyStockSearch:		nil,
		buyStock:			nil,
		manageFriends:		friends.NewManageFriendsPage(nil),
		createStockList:	stocklist.NewCreateStockListPage(nil),
		viewMyStockList: 	stocklist.NewViewStockLists(nil),
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
		m.mainMenu = menu.(*shared.MainMenuModel)

		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.mainMenu.Selected >= 0 && m.mainMenu.Selected < len(m.mainMenu.Options) {
			option := m.mainMenu.Options[m.mainMenu.Selected]
			if m.mainMenu.Confirmed {
				m.mainMenu.Confirmed = false
				switch option {
				case "Login":
					m.state = LoginState
					m.login = tuiauth.NewLogin()
				case "Register":
					m.state = RegisterState
					m.register = tuiauth.NewRegister()
				case "Configure":
					m.state = ConfigureState
					m.configure = shared.NewConfigure()
				case "Quit":
					return m, tea.Quit
				}
			}
		}
		return m, cmd

	case LoginState:
		login, cmd := m.login.Update(msg)
		m.login = login.(*tuiauth.LoginModel)
		// Check if user successfully logged in
		if m.login.GetUser() != nil {
			m.currentUser = *m.login.GetUser()
			m.state = HomePageState
			m.homepage = shared.NewHomePage(m.login.GetUser()) // reset homepage
			m.login = tuiauth.NewLogin() // reset login form
		}
		// Go back to main view from Login view
		if m.login.BackPressed {
			m.login.BackPressed = false
			m.state = MainMenuState
		}
		return m, cmd

	case RegisterState:
		register, cmd := m.register.Update(msg)
		m.register = register.(*tuiauth.RegisterModel)
		// Check if user successfully registered
		if m.register.GetUser() != nil {
			m.currentUser = *m.register.GetUser()
			m.state = HomePageState
			m.homepage = shared.NewHomePage(m.register.GetUser()) // reset homepage
			m.register = tuiauth.NewRegister() // reset register form
		}
		// Go back to main view from register view
		if m.register.BackPressed {
			m.register.BackPressed = false
			m.state = MainMenuState
		}
		return m, cmd

	case ConfigureState:
		configure, cmd := m.configure.Update(msg)
		m.configure = configure.(*shared.ConfigureModel)

		// Go back to home view from configure view
		if m.configure.BackPressed {
			m.configure.BackPressed = false
			m.state = MainMenuState
		}
		return m, cmd

	case HomePageState:
		homepage, cmd := m.homepage.Update(msg)
		m.homepage = homepage.(*shared.HomePageModel)

		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.homepage.Selected >= 0 && m.homepage.Selected < len(m.homepage.Options) {
			option := m.homepage.Options[m.homepage.Selected]
			if m.homepage.Confirmed {
				m.homepage.Confirmed = false
				switch option {
				case "My Portfolios":
					m.state = PortfolioState
					m.portfolio = portfolio.NewPortfolioPage()
				case "My Stock Lists":
					m.state = StockListState
					m.stockList = stocklist.NewStockListPage(m.homepage.GetUser())
				case "Stock Data & Analysis":
					m.state = StockAnalysisState
					m.stockAnalysis = stock.NewStockAnalysisPage()
				case "Friends & Social":
					m.state = SocialState
					m.social = shared.NewSocialPage(m.homepage.GetUser())
				}
			}
		}

		// Go back to main menu (logout) from HomePage
		if m.homepage.BackPressed {
			m.homepage.BackPressed = false
			m.state = MainMenuState
			m.currentUser = auth.User{} // clear current user
		}
		return m, cmd
	case PortfolioState:
		portfolioModel, cmd := m.portfolio.Update(msg)
		m.portfolio = portfolioModel.(*portfolio.PortfolioModel)

		// Check if user pressed Enter to select an option
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			option := m.portfolio.GetSelectedOption()
			userID := int(m.currentUser.UserID)
			if option == 0 {
				// View Portfolios selected
				m.state = ViewPortfoliosState
				m.viewPortfolios = portfolio.NewViewPortfoliosPageWithUserID(userID)
				cmd = m.viewPortfolios.Init()
			} else if option == 1 {
				// Create Portfolio selected
				m.state = CreatePortfolioState
				m.createPortfolio = portfolio.NewCreatePortfolioPageWithUserID(userID)
			}
		}

		// Go back to homepage from portfolio page
		if m.portfolio.BackPressed {
			m.portfolio.BackPressed = false
			m.state = HomePageState
		}
		return m, cmd
	case ViewPortfoliosState:
		viewPortfolios, cmd := m.viewPortfolios.Update(msg)
		m.viewPortfolios = viewPortfolios.(*portfolio.ViewPortfoliosModel)

		// Check if a portfolio was selected
		if portfolioMsg, ok := msg.(portfolio.PortfolioSelectedMsg); ok {
			// View Specific Portfolio selected
			m.state = ViewSpecPortfolioState
			m.viewSpecPortfolio = portfolio.NewViewSpecPortfolioPageWithUserID(int(m.currentUser.UserID), portfolioMsg.Portfolio.PortfolioID)
			// Pass the message to the new model to populate it
			return m, func() tea.Msg {
				return portfolioMsg
			}
		}

		// Go back to portfolio page from view portfolios page
		if m.viewPortfolios.BackPressed {
			m.viewPortfolios.BackPressed = false
			m.state = PortfolioState
			m.portfolio = portfolio.NewPortfolioPage()
		}
		return m, cmd

	case ViewSpecPortfolioState:
		viewSpecPortfolio, cmd := m.viewSpecPortfolio.Update(msg)
		m.viewSpecPortfolio = viewSpecPortfolio.(*portfolio.ViewSpecPortfolioModel)

		// Handle option selection from specific portfolio view
		if m.viewSpecPortfolio.OptionSelected != "" {
			option := m.viewSpecPortfolio.OptionSelected
			m.viewSpecPortfolio.OptionSelected = "" // Reset

			switch option {
			case "View Holdings":
				m.state = ViewHoldingsState
				portfolioID := m.viewSpecPortfolio.PortfolioID
				m.viewHoldings = stock.NewViewHoldingsPageWithPortfolioID(portfolioID)
				return m, m.viewHoldings.Init()
			case "Buy Stock":
				m.state = BuyStockSearchState
				userID := int(m.currentUser.UserID)
				portfolioID := m.viewSpecPortfolio.PortfolioID
				cashAccount := m.viewSpecPortfolio.Portfolio.CashAccount
				m.buyStockSearch = stock.NewBuyStockSearchPageWithPortfolio(userID, portfolioID, cashAccount)
				return m, m.buyStockSearch.Init()
			}
		}

		// Go back to view portfolios page from specific portfolio view page
		if m.viewSpecPortfolio.BackPressed {
			m.viewSpecPortfolio.BackPressed = false
			m.state = ViewPortfoliosState
			userID := int(m.currentUser.UserID)
			m.viewPortfolios = portfolio.NewViewPortfoliosPageWithUserID(userID)
			cmd = m.viewPortfolios.Init()
		}
		return m, cmd
	case CreatePortfolioState:
		createPortfolio, cmd := m.createPortfolio.Update(msg)
		m.createPortfolio = createPortfolio.(*portfolio.CreatePortfolioModel)
		// Go back to portfolio page from create portfolio page
		if m.createPortfolio.BackPressed {
			m.createPortfolio.BackPressed = false
			m.state = PortfolioState
			m.portfolio = portfolio.NewPortfolioPage()
		}
		return m, cmd
	case StockListState:
		stockList, cmd := m.stockList.Update(msg)
		m.stockList = stockList.(*stocklist.StockListModel)

		if m.stockList.Selected >= 0 && m.stockList.Selected < len(m.stockList.Options) {
			option := m.stockList.Options[m.stockList.Selected]
			if m.stockList.Confirmed {
				m.stockList.Confirmed = false
				switch option {
				case "Create New Stock List":
					m.state = CreateStockListState
					m.createStockList = stocklist.NewCreateStockListPage(m.stockList.GetUser())
				case "My Stock Lists":
					m.state = ViewStockListsState
					m.viewMyStockList = stocklist.NewViewStockLists(m.stockList.GetUser())
					cmd = m.viewMyStockList.Init()
				}
			}
		}
		// Go back to homepage from stock list page
		if m.stockList.BackPressed {
			m.stockList.BackPressed = false
			m.state = HomePageState
		}
		return m, cmd
	case CreateStockListState:
		createStockList, cmd := m.createStockList.Update(msg)
		m.createStockList = createStockList.(*stocklist.CreateStockListModel)
		// Go back to main stock list page from create stock list page
		if m.createStockList.BackPressed {
			m.createStockList.BackPressed = false
			m.state = StockListState
		}
		return m, cmd

	case ViewStockListsState:
		viewMyStockList, cmd := m.viewMyStockList.Update(msg)
		m.viewMyStockList = viewMyStockList.(*stocklist.ViewStockListsModel)
		// Go back to main stock list page from create stock list page
		if m.viewMyStockList.BackPressed {
			m.viewMyStockList.BackPressed = false
			m.state = StockListState
		}
		return m, cmd

	case StockAnalysisState:
		stockAnalysis, cmd := m.stockAnalysis.Update(msg)
		m.stockAnalysis = stockAnalysis.(*stock.StockAnalysisModel)

		// Handle menu option selection
		if m.stockAnalysis.Selected >= 0 && m.stockAnalysis.Selected < len(m.stockAnalysis.Options) {
			option := m.stockAnalysis.Options[m.stockAnalysis.Selected]
			if m.stockAnalysis.Confirmed {
				m.stockAnalysis.Confirmed = false
				switch option {
				case "View Current Stocks":
					m.state = CurrentStocksState
					m.currentStocks = stock.NewCurrentStocksPage()
					// Return Init command to fetch stocks data
					cmd = m.currentStocks.Init()
				case "Search Stock":
					m.state = SearchStockState
					m.searchStock = stock.NewSearchStockPage()
				}
			}
		}

		// Go back to homepage from Stock Data & Analysis page
		if m.stockAnalysis.BackPressed {
			m.stockAnalysis.BackPressed = false
			m.state = HomePageState
		}
		return m, cmd
	case SocialState:
		social, cmd := m.social.Update(msg)
		m.social = social.(*shared.SocialModel)

		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.social.Selected >= 0 && m.social.Selected < len(m.social.Options) {
			option := m.social.Options[m.social.Selected]
			if m.social.Confirmed {
				m.social.Confirmed = false
				switch option {
				case "Send Friend Request":
					m.state = SendFriReqState
					m.sendFriReq = friends.NewSendFriReqPage(m.social.GetUser())
				case "View Friends Requests":
					m.state = ViewFriReqState
					m.viewFriReq = friends.NewViewFriReqPage(m.social.GetUser())
				case "Manage Friends":
					m.state = ManageFriendsState
					m.manageFriends = friends.NewManageFriendsPage(m.social.GetUser())
					cmd = m.manageFriends.Init()
				}
			}
		}
		// Go back to homepage from social page
		if m.social.BackPressed {
			m.social.BackPressed = false
			m.state = HomePageState
		}
		return m, cmd

	case CurrentStocksState:
		currentStocks, cmd := m.currentStocks.Update(msg)
		m.currentStocks = currentStocks.(*stock.CurrentStocksModel)
		// Go back to stock analysis page from current stocks page
		if m.currentStocks.BackPressed {
			m.currentStocks.BackPressed = false
			m.state = StockAnalysisState
			m.stockAnalysis = stock.NewStockAnalysisPage()
		}
		return m, cmd
	case SearchStockState:
		searchStock, cmd := m.searchStock.Update(msg)
		m.searchStock = searchStock.(*stock.SearchStockModel)
		// Check if user confirmed search
		if m.searchStock.Confirmed {
			m.searchStock.Confirmed = false
			symbol := m.searchStock.GetSymbol()
			m.state = StockDetailsState
			m.stockDetails = stock.NewStockDetailsPage(symbol)
			// Return Init command to fetch historical data
			cmd = m.stockDetails.Init()
		}
		// Go back to stock analysis page
		if m.searchStock.BackPressed {
			m.searchStock.BackPressed = false
			m.state = StockAnalysisState
			m.searchStock = stock.NewSearchStockPage()
		}
		return m, cmd
	case StockDetailsState:
		stockDetails, cmd := m.stockDetails.Update(msg)
		m.stockDetails = stockDetails.(*stock.StockDetailsModel)
		// Go back to search stock page
		if m.stockDetails.BackPressed {
			m.stockDetails.BackPressed = false
			m.state = SearchStockState
			m.searchStock = stock.NewSearchStockPage()
		}
		return m, cmd
	case SendFriReqState:
		sendFriReq, cmd := m.sendFriReq.Update(msg)
		m.sendFriReq = sendFriReq.(*friends.SendFriReqModel)
		// Go back to friend and social from send friend request page
		if m.sendFriReq.BackPressed {
			m.sendFriReq.BackPressed = false
			m.state = SocialState
		}
		return m, cmd
	case ViewFriReqState:
		viewFriReq, cmd := m.viewFriReq.Update(msg)
		m.viewFriReq = viewFriReq.(*friends.ViewFriReqModel)
		// Go back to friend and social from view friend request page
		if m.viewFriReq.BackPressed {
			m.viewFriReq.BackPressed = false
			m.state = SocialState
		}
		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.viewFriReq.Selected >= 0 && m.viewFriReq.Selected < len(m.viewFriReq.Options) {
			option := m.viewFriReq.Options[m.viewFriReq.Selected]
			if m.viewFriReq.Confirmed {
				m.viewFriReq.Confirmed = false
				switch option {
				case "Incoming Requests (Accept / Reject)":
					m.state = IncFriReqState
					m.incFriReq = friends.NewIncFriReqPage(m.viewFriReq.GetUser())
					cmd = m.incFriReq.Init()
				case "Outgoing Requests (Cancel)":
					m.state = OutFriReqState
					m.outFriReq = friends.NewOutFriReqPage(m.viewFriReq.GetUser())
					cmd = m.outFriReq.Init()
				}
			}
		}
		return m, cmd
	case IncFriReqState:
		incFriReq, cmd := m.incFriReq.Update(msg)
		m.incFriReq = incFriReq.(*friends.IncFriReqModel)
		// Go back to view friend request page from incoming view page
		if m.incFriReq.BackPressed {
			m.incFriReq.BackPressed = false
			m.state = ViewFriReqState
		}
		return m, cmd
	case OutFriReqState:
		outFriReq, cmd := m.outFriReq.Update(msg)
		m.outFriReq = outFriReq.(*friends.OutFriReqModel)
		// Go back to view friend request page from outgoing view page
		if m.outFriReq.BackPressed {
			m.outFriReq.BackPressed = false
			m.state = ViewFriReqState
		}
		return m, cmd
	case ViewHoldingsState:
		viewHoldings, cmd := m.viewHoldings.Update(msg)
		m.viewHoldings = viewHoldings.(*stock.ViewHoldingsModel)

		// Go back to specific portfolio view from view holdings
		if m.viewHoldings.BackPressed {
			m.viewHoldings.BackPressed = false
			m.state = ViewSpecPortfolioState
		}
		return m, cmd

	case BuyStockSearchState:
		buyStockSearch, cmd := m.buyStockSearch.Update(msg)
		m.buyStockSearch = buyStockSearch.(*stock.BuyStockSearchModel)

		// Check if a stock was selected
		if stockMsg, ok := msg.(stock.StockSelectedForBuyMsg); ok {
			m.state = BuyStockState
			userID := int(m.currentUser.UserID)
			portfolioID := m.buyStockSearch.PortfolioID
			cashAccount := m.buyStockSearch.CashAccount
			m.buyStock = stock.NewBuyStockPageWithStock(userID, portfolioID, stockMsg.Stock, cashAccount)
		}

		// Go back to specific portfolio view from buy stock search
		if m.buyStockSearch.BackPressed {
			m.buyStockSearch.BackPressed = false
			m.state = ViewSpecPortfolioState
		}
		return m, cmd
	case BuyStockState:
		buyStock, cmd := m.buyStock.Update(msg)
		m.buyStock = buyStock.(*stock.BuyStockModel)

		// Handle confirmed purchase
		if m.buyStock.Confirmed {
			m.buyStock.Confirmed = false
			// Execute the buy transaction asynchronously
			return m, func() tea.Msg {
				db := database.New()
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				// Convert portfolioID from string to int
				portfolioID := 0
				fmt.Sscanf(m.buyStock.PortfolioID, "%d", &portfolioID)

				err := db.BuyStock(
					ctx,
					portfolioID,
					m.buyStock.Stock.Symbol,
					m.buyStock.GetQuantity(),
					m.buyStock.Stock.Price,
				)

				if err != nil {
					return stock.BuyStockCompletedMsg{
						Success: false,
						Message: "✗ " + err.Error(),
					}
				}

				return stock.BuyStockCompletedMsg{
					Success: true,
					Message: "✓ Stock purchase completed!",
				}
			}
		}

		// Handle buy completion message
		if completedMsg, ok := msg.(stock.BuyStockCompletedMsg); ok {
			if completedMsg.Success {
				// Purchase successful, go back to portfolio view and refresh data
				m.state = ViewSpecPortfolioState
				return m, m.viewSpecPortfolio.RefreshPortfolio()
			} else {
				// Show error message in the buy stock page
				m.buyStock.Error = completedMsg.Message
			}
		}

		// Go back to stock search from buy stock page
		if m.buyStock.BackPressed {
			m.buyStock.BackPressed = false
			m.state = BuyStockSearchState
		}

		return m, cmd
	case ManageFriendsState:
		manageFriends, cmd := m.manageFriends.Update(msg)
		m.manageFriends = manageFriends.(*friends.ManageFriendsModel)
		// Go back to friend & social page from manage friends page
		if m.manageFriends.BackPressed {
			m.manageFriends.BackPressed = false
			m.state = SocialState
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
	case ViewSpecPortfolioState:
		return m.viewSpecPortfolio.View()
	case ViewHoldingsState:
		return m.viewHoldings.View()
	case BuyStockSearchState:
		return m.buyStockSearch.View()
	case BuyStockState:
		return m.buyStock.View()
	case ManageFriendsState:
		return m.manageFriends.View()
	case CreateStockListState:
		return m.createStockList.View()
	case ViewStockListsState:
		return m.viewMyStockList.View()
	}
	return ""
}

// StartApp starts the interactive CLI app
func StartApp() error {
	p := tea.NewProgram(NewAppModel())
	_, err := p.Run()
	return err
}

