package tui

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

type FooterModel struct {
	height, width int
	h             help.Model
	currentKeymap help.KeyMap
}

func NewFooterModel() FooterModel {
	return FooterModel{
		height: 1,
	}
}

func (f FooterModel) Init() tea.Cmd {
	return nil
}

func (f FooterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// the width should already be adjusted for the appended on borders
		f.width = msg.Width
		f.h.Width = msg.Width
	}
	return f, nil
}

func (f FooterModel) View() string {
	return f.h.View(f.currentKeymap)
}

func (f FooterModel) Height() int {
	return f.height
}

func (f FooterModel) Width() int {
	return f.width
}
