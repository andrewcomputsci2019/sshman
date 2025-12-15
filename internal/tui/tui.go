package tui

import (
	"andrew/sshman/internal/sqlite"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type appState int

const (
	viewMode   appState = iota // user on main screen
	editMode                   // user is in edit mode editing a host
	wizardMode                 // user is in entry wizard mode adding a host
	navigateMode
)

type updateHostsMessage struct {
	host sqlite.Host
}

type newHostsMessage struct {
	host sqlite.Host
}

type userAddHostMessage struct{}

type AppModel struct {
	width, height int // this constitutes the entire terminal size
	// app components
	footer     FooterModel
	header     HeaderModel
	hostsModel HostsPanelModel
	db         *sqlite.HostDao
}

// todo implement model func

func (a AppModel) Init() tea.Cmd {
	//TODO implement me
	panic("implement me")
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//TODO implement me
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		//todo call update on submodules to have them update their height and width but with adjusted bounds
	}

	panic("implement me")
}

func (a AppModel) View() string {
	header := lipgloss.NewStyle().Width(a.width).
		Height(a.header.Height()).
		Render(a.header.View())
	footer := lipgloss.NewStyle().
		Width(a.width).
		Height(a.footer.Height()).
		Render(a.footer.View())
	// todo add host model to this
	return lipgloss.JoinVertical(lipgloss.Left, header, footer)
}
