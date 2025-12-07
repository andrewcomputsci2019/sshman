package tui

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

type FooterModel struct {
	height int
	help.Model
}

func (f FooterModel) Init() tea.Cmd {
	//TODO implement me
	panic("implement me")
}

func (f FooterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//TODO implement me
	panic("implement me")
}

func (f FooterModel) View() string {
	//TODO implement me
	panic("implement me")
}

func (h FooterModel) Height() int {
	return h.height
}
