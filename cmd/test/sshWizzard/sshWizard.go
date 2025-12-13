package main

import (
	"fmt"
	"log"
	"strings"

	"andrew/sshman/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

const wizardSubmissionMsgType = "tui.newHostsMessage"

// wizardHarnessModel is a thin Bubble Tea wrapper that embeds the actual
// sshWizard model so that we can quickly exercise the wizard UI on its own.
// This is handy for manual testing while iterating on the wizard logic.
type wizardHarnessModel struct {
	wizard tui.WizardViewModel
	status string
}

func newWizardHarnessModel() wizardHarnessModel {
	return wizardHarnessModel{
		wizard: tui.NewWizardViewModel(),
		status: "Fill out the form and press Enter on \"Confirm\" to submit. Press q to exit.",
	}
}

func (m wizardHarnessModel) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, m.wizard.Init())
}

func (m wizardHarnessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch key := msg.(type) {
	case tea.KeyMsg:
		if key.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	if submission := tryExtractWizardSubmission(msg); submission != "" {
		m.status = fmt.Sprintf("Submission: %s", submission)
		return m, nil
	}

	updated, cmd := m.wizard.Update(msg)
	if wiz, ok := updated.(tui.WizardViewModel); ok {
		m.wizard = wiz
	}
	return m, cmd
}

func (m wizardHarnessModel) View() string {
	b := &strings.Builder{}
	b.WriteString("SSH Wizard Harness\n")
	b.WriteString(m.status)
	b.WriteString("\n\n")
	b.WriteString(m.wizard.View())
	return b.String()
}

func tryExtractWizardSubmission(msg tea.Msg) string {
	if fmt.Sprintf("%T", msg) != wizardSubmissionMsgType {
		return ""
	}
	return fmt.Sprintf("%+v", msg)
}

func main() {
	program := tea.NewProgram(newWizardHarnessModel(), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatalf("failed to run ssh wizard harness: %v", err)
	}
}
