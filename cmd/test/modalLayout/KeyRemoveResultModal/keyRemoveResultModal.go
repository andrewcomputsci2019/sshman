package main

import (
	"andrew/sshman/internal/config"
	"andrew/sshman/internal/sqlite"
	"andrew/sshman/internal/tui"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type showResultModalMsg struct{}

type resultModalHarness struct {
	app tui.AppModel
}

func newRemoveResultHarness(app tui.AppModel) resultModalHarness {
	return resultModalHarness{app: app}
}

func (m resultModalHarness) Init() tea.Cmd {
	return tea.Batch(
		m.app.Init(),
		tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg {
			return showResultModalMsg{}
		}),
	)
}

func (m resultModalHarness) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	switch msg.(type) {
	case showResultModalMsg:
		m.app.ShowKeyRemoveResultModal(nil)
		return m, nil
	}

	updated, cmd := m.app.Update(msg)
	if app, ok := updated.(tui.AppModel); ok {
		m.app = app
	}
	return m, cmd
}

func (m resultModalHarness) View() string {
	return m.app.View()
}

func main() {
	db, err := sqlite.CreateAndLoadDB(":memory:")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	hostDao := sqlite.NewHostDao(db)
	cfg := config.Config{}
	cfg.Ssh.KeyPath = os.TempDir()
	cfg.StorageConf.WriteThrough = new(bool)
	*cfg.StorageConf.WriteThrough = false
	cfg.DevMode = true
	app := tui.NewAppModel([]sqlite.Host{}, hostDao, cfg)
	program := tea.NewProgram(newRemoveResultHarness(app), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Printf("err: %s", err)
	}
}
