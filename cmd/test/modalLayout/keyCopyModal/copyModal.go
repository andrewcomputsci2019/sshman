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

type showKeyCopyModalMsg struct{}

type copyModalHarness struct {
	app tui.AppModel
}

func newCopyModalHarness(app tui.AppModel) copyModalHarness {
	return copyModalHarness{app: app}
}

func (m copyModalHarness) Init() tea.Cmd {
	return tea.Batch(
		m.app.Init(),
		tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg {
			return showKeyCopyModalMsg{}
		}),
	)
}

func (m copyModalHarness) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	switch msg.(type) {
	case showKeyCopyModalMsg:
		m.app.ShowKeyCopyModal(nil)
		return m, nil
	}

	updated, cmd := m.app.Update(msg)
	if app, ok := updated.(tui.AppModel); ok {
		m.app = app
	}
	return m, cmd
}

func (m copyModalHarness) View() string {
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
	program := tea.NewProgram(newCopyModalHarness(app), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Printf("err: %s", err)
	}
}
