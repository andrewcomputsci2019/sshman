package tui

import (
	"andrew/sshman/internal/config"
	"andrew/sshman/internal/sqlite"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type TableKeyBinds struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Edit      key.Binding
	Add       key.Binding
	Delete    key.Binding
	Select    key.Binding
	CycleView key.Binding
}

func (t TableKeyBinds) ShortHelp() []key.Binding {
	return []key.Binding{t.Up, t.Down, t.Left, t.Right, t.Edit, t.Add, t.Delete, t.Select, t.CycleView}
}

func (t TableKeyBinds) FullHelp() [][]key.Binding {
	binds := make([][]key.Binding, 0)
	binds = append(binds, []key.Binding{t.Up, t.Down, t.Left, t.Right})
	binds = append(binds, []key.Binding{t.Edit, t.Add, t.Delete})
	binds = append(binds, []key.Binding{t.Select, t.CycleView})
	return binds
}

type InfoViewKeyBinds struct {
	Up             key.Binding // j
	Down           key.Binding // k
	Next           key.Binding // tab
	Prev           key.Binding // shift tab
	CollapseToggle key.Binding // alt-c
	Save           key.Binding // ctrl-s only works in edit mode
	ChangeView     key.Binding // ctrl-w
	CancelView     key.Binding // this is like exit and go back to table focus
}

func (i InfoViewKeyBinds) ShortHelp() []key.Binding {
	return []key.Binding{i.Up, i.Down, i.Next, i.Prev, i.CollapseToggle, i.Save, i.ChangeView, i.CancelView}
}

func (i InfoViewKeyBinds) FullHelp() [][]key.Binding {
	binds := make([][]key.Binding, 0)
	binds = append(binds, []key.Binding{i.Up, i.Down, i.Next, i.Prev})
	binds = append(binds, []key.Binding{i.Save, i.CancelView})
	binds = append(binds, []key.Binding{i.CollapseToggle, i.ChangeView})
	return binds
}

var tableKeyMap TableKeyBinds = TableKeyBinds{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "up")),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "down")),
	Left: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("h/←", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("l/→", "right")),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit")),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add")),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete")),
	Select: key.NewBinding(key.WithKeys("enter"),
		key.WithHelp("enter", "connect to host")),
	CycleView: key.NewBinding(
		key.WithKeys("ctrl+w"),
		key.WithHelp("ctrl+w", "cycle views")),
}

var infoPanelKeyMap InfoViewKeyBinds = InfoViewKeyBinds{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "up")),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "down")),
	Next: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next")),
	Prev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev")),
	CollapseToggle: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "collapse")),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save")),
	ChangeView: key.NewBinding(
		key.WithKeys("alt+w"),
		key.WithHelp("alt+w", "change focus view")),
	CancelView: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "exit/cancel")),
}

type HostsModel struct {
	table         table.Model
	cfg           config.Config
	width, height int
}

func (h HostsModel) Init() tea.Cmd {
	//TODO implement me
	panic("implement me")
}

func (h HostsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//TODO implement me
	panic("implement me")
}

func (h HostsModel) View() string {
	//TODO implement me
	panic("implement me")
}

type kvInputModel struct {
	key           textinput.Model
	val           textinput.Model
	Editable      bool
	neverEditable bool
}

func newKvInputModel(key string, val string, initState bool, neverEdit bool) kvInputModel {
	kv := kvInputModel{
		key:           textinput.New(),
		val:           textinput.New(),
		Editable:      initState,
		neverEditable: neverEdit,
	}
	kv.key.Placeholder = "key"
	kv.val.Placeholder = "value"
	kv.key.SetValue(key)
	kv.val.SetValue(val)
	return kv
}

type HostsInfoModel struct {
	hostOptions             []kvInputModel
	optionsScrollPane       viewport.Model
	host                    string
	previewOptionScrollPane viewport.Model
	hostNotes               textarea.Model
	currentEditHost         sqlite.Host
	HostPreviewString       string
	width, height           int
}

func (h HostsInfoModel) Init() tea.Cmd {
	//TODO implement me
	panic("implement me")
}

func (h HostsInfoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//TODO implement me
	panic("implement me")
}

func (h HostsInfoModel) View() string {
	// todo tune cutoff for auto collapse of preview
	if h.height < 20 {
		// dont render preview
	} else {

	}
	//TODO implement me
	panic("implement me")
}

const (
	focusTable     = iota
	focusInfoPanel = iota
)

type HostsPanelModel struct {
	table           HostsModel
	infoPanel       HostsInfoModel
	focus           int
	data            []sqlite.Host
	tableGrowthBias float32 // value of .70 means table should take 70% of total width
}

func (h HostsPanelModel) ShortHelp() []key.Binding {
	if h.focus == focusTable {
		return tableKeyMap.ShortHelp()
	} else {
		return infoPanelKeyMap.ShortHelp()
	}
}

func (h HostsPanelModel) FullHelp() [][]key.Binding {
	if h.focus == focusTable {
		return tableKeyMap.FullHelp()
	} else {
		return infoPanelKeyMap.FullHelp()
	}
}

func (h HostsPanelModel) Init() tea.Cmd {
	return nil
}

func (h HostsPanelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width < 80 { // go into vertical render mode

		} else {

		}
		h.table.width = int(float32(msg.Width)*h.tableGrowthBias) - 2
		h.table.height = msg.Height - 2
		h.infoPanel.width = msg.Width - h.table.width - 1 + 2
		h.infoPanel.height = msg.Height - 2
	}
	//TODO implement me
	panic("implement me")
	return h, nil
}

func (h HostsPanelModel) View() string {
	if h.focus == focusTable {
		return h.table.View()
	} else {
		return h.infoPanel.View()
	}
}
