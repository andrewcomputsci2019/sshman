package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type HeaderModel struct {
	BuildInfo     string
	height, width int
	numberOfHost  uint
	buildVersion  string
	buildDate     string
	buildArch     string
	programName   string
}

func NewHeaderModel(numberOfHost uint, buildVersion, buildDate, buildArch, programName string) HeaderModel {
	return HeaderModel{
		height:       1,
		numberOfHost: numberOfHost,
		buildVersion: buildVersion,
		buildDate:    buildDate,
		buildArch:    buildArch,
		programName:  programName,
	}
}

func (h HeaderModel) Init() tea.Cmd {
	//TODO implement me
	panic("implement me")
}

func (h HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
	}
	return h, nil
}

func (h HeaderModel) View() string {
	//TODO implement me, render build info, program name, total host, etc
	panic("implement me")
}

func (h HeaderModel) Height() int {
	return h.height
}

func (h HeaderModel) Width() int {
	return h.width
}
