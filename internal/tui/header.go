package tui

import (
	"andrew/sshman/internal/buildInfo"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HeaderModel struct {
	height, width int
	numberOfHost  uint
	buildMajor    int
	buildMinor    int
	buildPatch    int
	buildDate     string
	buildOS       string
	buildArch     string
	programName   string
}

func NewHeaderModel(numberOfHost uint) HeaderModel {
	return HeaderModel{
		height:       1,
		numberOfHost: numberOfHost,
		buildDate:    buildInfo.BuildDate,
		buildArch:    buildInfo.BUILD_ARC,
		buildOS:      buildInfo.BUILD_OS,
		buildMajor:   buildInfo.BuildMajor,
		buildMinor:   buildInfo.BuildMinor,
		buildPatch:   buildInfo.BuildPatch,
		programName:  buildInfo.ProgramName,
	}
}

func (h HeaderModel) Init() tea.Cmd {
	return nil
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
	separator := "⏺"
	numHostString := strconv.Itoa(int(h.numberOfHost))
	majorStyle := lipgloss.NewStyle().Background(lipgloss.Color("46")).Foreground(lipgloss.Color("#000"))
	minorVersion := lipgloss.NewStyle().Background(lipgloss.Color("51")).Foreground(lipgloss.Color("#000"))
	patchVersion := lipgloss.NewStyle().Background(lipgloss.Color("39")).Foreground(lipgloss.Color("#000"))
	base := lipgloss.NewStyle().Foreground(lipgloss.Color("#fff"))
	render := lipgloss.JoinHorizontal(lipgloss.Left,
		base.Render("🛠️\t"+h.programName+" "+separator+" "),
		base.Render("Build: "),
		majorStyle.Render(strconv.Itoa(h.buildMajor)+"."),
		minorVersion.Render(strconv.Itoa(h.buildMinor)+"."),
		patchVersion.Render(strconv.Itoa(h.buildPatch)),
		base.Render(
			" Build Date: "+h.buildDate+
				" Instruction Set: "+h.buildArch+
				" "+separator+" OS: "+h.buildOS+
				" "+separator+" Total Hosts Managed: "+numHostString,
		),
	)
	return lipgloss.NewStyle().
		Width(h.width).
		MaxWidth(h.width).
		Render(render)
}

func (h HeaderModel) Height() int {
	return h.height
}

func (h HeaderModel) Width() int {
	return h.width
}
