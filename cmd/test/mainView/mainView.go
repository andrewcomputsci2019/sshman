package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"andrew/sshman/internal/config"
	"andrew/sshman/internal/sqlite"
	"andrew/sshman/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	hostConnectMsgType = "tui.connectHostMessage"
	hostUpdateMsgType  = "tui.updateHostsMessage"
	hostNewMsgType     = "tui.newHostsMessage"
)

// hostPanelHarnessModel wraps the HostsPanelModel so we can exercise it in
// isolation from the rest of the TUI while iterating on the UI.
type hostPanelHarnessModel struct {
	panel  tui.HostsPanelModel
	status string
}

func newHostPanelHarnessModel() hostPanelHarnessModel {
	cfg := config.Config{
		EnablePing: true,
		Ssh: config.SSH{
			KeyOnly: true,
		},
	}
	return hostPanelHarnessModel{
		panel: tui.NewHostsPanelModel(cfg, demoHosts()),
		status: "Use j/k to move, e to edit, ctrl+w to toggle focus. " +
			"Press q to exit.",
	}
}

func (m hostPanelHarnessModel) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, m.panel.Init())
}

func (m hostPanelHarnessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	updated, cmd := m.panel.Update(msg)
	if panel, ok := updated.(tui.HostsPanelModel); ok {
		m.panel = panel
	}

	if status, handled := interpretHarnessEvent(msg); handled {
		m.status = status
	}
	return m, cmd
}

func (m hostPanelHarnessModel) View() string {
	var b strings.Builder
	b.WriteString("Hosts Panel Harness\n")
	b.WriteString(m.status)
	b.WriteString("\n\n")
	b.WriteString(m.panel.View())
	return b.String()
}

func interpretHarnessEvent(msg tea.Msg) (string, bool) {
	description := fmt.Sprintf("%+v", msg)
	switch fmt.Sprintf("%T", msg) {
	case hostConnectMsgType:
		return fmt.Sprintf("Connect requested: %s", description), true
	case hostUpdateMsgType:
		return fmt.Sprintf("Host updated: %s", description), true
	case hostNewMsgType:
		return fmt.Sprintf("New host added: %s", description), true
	default:
		return "", false
	}
}

func demoHosts() []sqlite.Host {
	now := time.Now()
	return []sqlite.Host{
		{
			Host:           "prod-web-01",
			CreatedAt:      now.Add(-72 * time.Hour),
			UpdatedAt:      timePtr(now.Add(-2 * time.Hour)),
			LastConnection: timePtr(now.Add(-45 * time.Minute)),
			Notes:          "Primary production web node",
			Tags:           []string{"prod", "web"},
			Options: []sqlite.HostOptions{
				{Key: "Hostname", Value: "prod-web-01.internal"},
				{Key: "User", Value: "deploy"},
				{Key: "Port", Value: "22"},
				{Key: "ForwardAgent", Value: "yes"},
			},
		},
		{
			Host:           "staging-db",
			CreatedAt:      now.Add(-48 * time.Hour),
			UpdatedAt:      timePtr(now.Add(-4 * time.Hour)),
			LastConnection: timePtr(now.Add(-90 * time.Minute)),
			Notes:          "Staging database with non-standard ssh port",
			Tags:           []string{"staging", "db"},
			Options: []sqlite.HostOptions{
				{Key: "Hostname", Value: "staging-db.internal"},
				{Key: "User", Value: "dbadmin"},
				{Key: "Port", Value: "2202"},
				{Key: "IdentityFile", Value: "~/.ssh/staging_db"},
			},
		},
		{
			Host:      "dev-scratch",
			CreatedAt: now.Add(-24 * time.Hour),
			Notes:     "Playground box for testing host editing workflow",
			Tags:      []string{"dev", "scratch"},
			Options: []sqlite.HostOptions{
				{Key: "Hostname", Value: "192.168.0.96"},
				{Key: "User", Value: "admin"},
				{Key: "ProxyJump", Value: "bastion"},
			},
		},
		{
			Host:      "ping-test",
			CreatedAt: now.Add(-26 * time.Hour),
			Notes:     "Test Ping Command",
			Tags:      []string{"ping"},
			Options: []sqlite.HostOptions{
				{Key: "User", Value: "test"},
				{Key: "Hostname", Value: "localhost"},
			},
		},
	}
}

func timePtr(t time.Time) *time.Time {
	ts := t
	return &ts
}

func main() {
	program := tea.NewProgram(newHostPanelHarnessModel(), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatalf("failed to run hosts panel harness: %v", err)
	}
}
