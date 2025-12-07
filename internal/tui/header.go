package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type HeaderModel struct {
	BuildInfo string
	height    int
}

func (h HeaderModel) Init() tea.Cmd {
	//TODO implement me
	panic("implement me")
}

func (h HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//TODO implement me
	panic("implement me")
}

func (h HeaderModel) View() string {
	//TODO implement me
	panic("implement me")
}

func (h HeaderModel) Height() int {
	return h.height
}
