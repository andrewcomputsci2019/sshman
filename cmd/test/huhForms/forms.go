package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"andrew/sshman/internal/config"
	"andrew/sshman/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	keyGenResultMsgType     = "tui.keyGenResult"
	abortedKeyGenMsgType    = "tui.abortedKeyGenForm"
	keyRotateResultMsgType  = "tui.keyRotateRequest"
	abortedKeyRotateMsgType = "tui.abortedRotatedKeyForm"
)

type keyViewHarnessMode int

const (
	keyGenMode keyViewHarnessMode = iota
	keyRotateMode
)

// keyViewHarnessModel wraps the key view forms so they can be exercised
// without launching the full TUI app.
type keyViewHarnessModel struct {
	mode      keyViewHarnessMode
	keyGen    tui.KeyGenModel
	keyRotate tui.KeyRotateModel
	status    string
	keyPath   string
}

func newKeyViewHarnessModel() keyViewHarnessModel {
	cfg, keyPath := demoKeyConfig()
	keys := demoKeys(keyPath)
	model := keyViewHarnessModel{
		mode:      keyGenMode,
		keyGen:    tui.NewKeyGenModel("demo-host", cfg),
		keyRotate: tui.NewKeyRotateModel("demo-host", keys, cfg),
		status:    fmt.Sprintf("g: key gen, r: rotate, ctrl+c: exit. key path: %s", keyPath),
		keyPath:   keyPath,
	}
	return model
}

func (m keyViewHarnessModel) Init() tea.Cmd {
	if err := os.MkdirAll(m.keyPath, 0o700); err != nil {
		m.status = fmt.Sprintf("failed to create key path %s: %v", m.keyPath, err)
	}
	return tea.Batch(tea.EnterAltScreen, m.keyGen.Init(), m.keyRotate.Init())
}

func (m keyViewHarnessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "g":
			m.mode = keyGenMode
			m.status = fmt.Sprintf("key gen mode. key path: %s", m.keyPath)
		case "r":
			m.mode = keyRotateMode
			m.status = fmt.Sprintf("key rotate mode. key path: %s", m.keyPath)
		}
	}

	if status, handled := interpretKeyViewEvent(msg); handled {
		m.status = status
	}

	switch m.mode {
	case keyRotateMode:
		updated, cmd := m.keyRotate.Update(msg)
		if rotate, ok := updated.(tui.KeyRotateModel); ok {
			m.keyRotate = rotate
		}
		return m, cmd
	default:
		updated, cmd := m.keyGen.Update(msg)
		if gen, ok := updated.(tui.KeyGenModel); ok {
			m.keyGen = gen
		}
		return m, cmd
	}
}

func (m keyViewHarnessModel) View() string {
	var b strings.Builder
	b.WriteString("Key View Harness\n")
	b.WriteString(m.status)
	b.WriteString("\n\n")
	switch m.mode {
	case keyRotateMode:
		b.WriteString(m.keyRotate.View())
	default:
		b.WriteString(m.keyGen.View())
	}
	return b.String()
}

func interpretKeyViewEvent(msg tea.Msg) (string, bool) {
	switch fmt.Sprintf("%T", msg) {
	case keyGenResultMsgType:
		return fmt.Sprintf("Key gen result: %+v", msg), true
	case abortedKeyGenMsgType:
		return "Key gen form aborted", true
	case keyRotateResultMsgType:
		return fmt.Sprintf("Key rotate result: %+v", msg), true
	case abortedKeyRotateMsgType:
		return "Key rotate form aborted", true
	default:
		return "", false
	}
}

func demoKeyConfig() (config.Config, string) {
	keyPath := filepath.Join(os.TempDir(), "sshman-test-keys")
	return config.Config{
		Ssh: config.SSH{
			KeyPath:                    keyPath,
			AcceptableKeyGenAlgorithms: []string{config.RSA, config.ECDSA, config.ED25519},
		},
	}, keyPath
}

func demoKeys(keyPath string) []string {
	return []string{
		filepath.Join(keyPath, "rsa_demo-host_abcd12_20240101"),
		filepath.Join(keyPath, "ecdsa_demo-host_ef3412_20240101"),
		filepath.Join(keyPath, "ed25519_demo-host_aa9911_20240101"),
	}
}

func main() {
	program := tea.NewProgram(newKeyViewHarnessModel(), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatalf("failed to run key view harness: %v", err)
	}
}
