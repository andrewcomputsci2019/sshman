package tui

import (
	"andrew/sshman/internal/config"
	"andrew/sshman/internal/sqlite"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	hostColumnKey              = "host"
	hostHostnameColumnKey      = "hostname"
	hostLastConnectedColumnKey = "last_connected"
	hostPingColumnKey          = "ping"
	hostStatusColumnKey        = "status"
	hostRowPayloadKey          = "__host_payload"
)

const (
	infoViewMode = iota
	infoEditMode
)

const (
	minimumTableWidth       = 32
	minimumInfoWidth        = 28
	verticalLayoutThreshold = 80
	defaultTableBias        = 0.65
	defaultNotesHeight      = 5
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
		key.WithKeys("ctrl+w"),
		key.WithHelp("ctrl+w", "change focus view")),
	CancelView: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "exit/cancel")),
}

type HostsModel struct {
	table         table.Model
	cfg           config.Config
	width, height int
}

func NewHostsModel(cfg config.Config) HostsModel {
	columns := []table.Column{
		table.NewColumn(hostColumnKey, "Host", 20),
		table.NewColumn(hostHostnameColumnKey, "Hostname", 16),
		table.NewColumn(hostLastConnectedColumnKey, "Last Connected", 18),
		table.NewColumn(hostPingColumnKey, "Ping", 8),
		table.NewColumn(hostStatusColumnKey, "Status", 10),
	}
	tbl := table.New(columns).
		WithRows([]table.Row{}).
		WithMinimumHeight(5).
		WithRowStyleFunc(func(input table.RowStyleFuncInput) lipgloss.Style {
			if input.IsHighlighted {
				return lipgloss.NewStyle().Background(lipgloss.Color("#2E2E3E"))
			}
			if input.Index%2 == 0 {
				return lipgloss.NewStyle().Foreground(lipgloss.Color("#D1D5DB"))
			}
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#F3F4F6"))
		})
	return HostsModel{
		table: tbl,
		cfg:   cfg,
	}
}

func (h HostsModel) Init() tea.Cmd {
	return nil
}

func (h HostsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg.Type, tableKeyMap.Add) {
			// handle add event.
		}
	}
	h.table, cmd = h.table.Update(msg)
	return h, cmd
}

func (h HostsModel) View() string {
	return h.table.View()
}

func (h *HostsModel) setSize(width, height int) {
	if width < minimumTableWidth {
		width = minimumTableWidth
	}
	if height < 3 {
		height = 3
	}
	h.width = width
	h.height = height
	h.table = h.table.WithTargetWidth(width).WithMinimumHeight(height)
}

func (h *HostsModel) setFocused(focused bool) {
	h.table = h.table.Focused(focused)
}

func (h *HostsModel) setRows(rows []table.Row) {
	h.table = h.table.WithRows(rows)
}

func (h HostsModel) highlightedHost() *sqlite.Host {
	row := h.table.HighlightedRow()
	if payload, ok := row.Data[hostRowPayloadKey]; ok {
		if host, ok := payload.(*sqlite.Host); ok {
			return host
		}
	}
	return nil
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

func (k *kvInputModel) setWidth(total int) {
	if total < 10 {
		total = 10
	}
	keyWidth := max(12, total/3)
	valueWidth := max(12, total-keyWidth-4)
	k.key.Width = keyWidth
	k.val.Width = valueWidth
}

type HostsInfoModel struct {
	hostOptions             []kvInputModel
	optionsScrollPane       viewport.Model
	host                    string
	previewOptionScrollPane viewport.Model
	hostNotes               textarea.Model
	tagsInput               textinput.Model
	currentEditHost         sqlite.Host
	HostPreviewString       string
	width, height           int
	focused                 bool
	mode                    int
	selected                int
	previewCollapsed        bool
	pendingSave             bool
}

func NewHostsInfoModel() HostsInfoModel {
	optionsViewport := viewport.New(0, 0)
	previewViewport := viewport.New(0, 0)
	notes := textarea.New()
	notes.Placeholder = "Notes"
	notes.SetHeight(defaultNotesHeight)
	notes.Blur()
	tags := textinput.New()
	tags.Placeholder = "tag1,tag2"
	tags.Width = 20
	return HostsInfoModel{
		optionsScrollPane:       optionsViewport,
		previewOptionScrollPane: previewViewport,
		hostNotes:               notes,
		tagsInput:               tags,
		mode:                    infoViewMode,
		selected:                0,
	}
}

func (h HostsInfoModel) Init() tea.Cmd {
	return nil
}

func (h HostsInfoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !h.focused {
			return h, nil
		}
		switch {
		case key.Matches(msg, infoPanelKeyMap.CollapseToggle):
			h.previewCollapsed = !h.previewCollapsed
			return h, nil
		case key.Matches(msg, infoPanelKeyMap.Up), key.Matches(msg, infoPanelKeyMap.Prev):
			cmd := h.moveSelection(-1)
			return h, cmd
		case key.Matches(msg, infoPanelKeyMap.Down), key.Matches(msg, infoPanelKeyMap.Next):
			cmd := h.moveSelection(1)
			return h, cmd
		case key.Matches(msg, infoPanelKeyMap.Save) && h.mode == infoEditMode:
			updated := h.buildUpdatedHost()
			h.currentEditHost = updated
			h.HostPreviewString = buildHostPreview(updated)
			h.pendingSave = true
			h.mode = infoViewMode
			h.blurCurrentInput()
			return h, func() tea.Msg {
				return updateHostsMessage{host: updated}
			}
		}
	}
	if !h.focused {
		return h, nil
	}
	if h.mode == infoEditMode {
		if h.selectionIsTags() {
			var cmd tea.Cmd
			h.tagsInput, cmd = h.tagsInput.Update(msg)
			return h, cmd
		}
		if h.selectionIsNotes() {
			var cmd tea.Cmd
			h.hostNotes, cmd = h.hostNotes.Update(msg)
			return h, cmd
		}
		if h.selected >= 0 && h.selected < len(h.hostOptions) && !h.hostOptions[h.selected].neverEditable {
			val := h.hostOptions[h.selected].val
			var cmd tea.Cmd
			h.hostOptions[h.selected].val, cmd = val.Update(msg)
			return h, cmd
		}
	}
	return h, nil
}

func (h HostsInfoModel) View() string {
	if h.host == "" {
		return "No host selected"
	}
	title := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Host %s", h.host))
	var sections []string
	sections = append(sections, title)

	h.optionsScrollPane.SetContent(h.renderOptions())
	sections = append(sections, lipgloss.NewStyle().Bold(true).Render("Options"))
	sections = append(sections, h.optionsScrollPane.View())

	tagsLabel := "Tags"
	if h.focused && h.selectionIsTags() {
		tagsLabel = "> Tags"
	}
	sections = append(sections, lipgloss.NewStyle().Bold(true).Render(tagsLabel))
	sections = append(sections, h.renderTags())

	notesLabel := "Notes"
	if h.focused && h.selectionIsNotes() {
		notesLabel = "> Notes"
	}
	sections = append(sections, lipgloss.NewStyle().Bold(true).Render(notesLabel))
	sections = append(sections, h.renderNotes())

	if h.shouldRenderPreview() {
		h.previewOptionScrollPane.SetContent(h.HostPreviewString)
		sections = append(sections, lipgloss.NewStyle().Bold(true).Render("Preview"))
		sections = append(sections, h.previewOptionScrollPane.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (h *HostsInfoModel) setSize(width, height int) {
	if width < minimumInfoWidth {
		width = minimumInfoWidth
	}
	if height < 6 {
		height = 6
	}
	h.width = width
	h.height = height
	h.optionsScrollPane.Width = width
	h.optionsScrollPane.Height = max(4, height/2)
	h.previewOptionScrollPane.Width = width
	h.previewOptionScrollPane.Height = max(3, height-h.optionsScrollPane.Height-defaultNotesHeight-2)
	h.hostNotes.SetWidth(max(10, width-2))
	h.tagsInput.Width = max(10, width-2)
	h.hostNotes.SetHeight(max(defaultNotesHeight, height/4))
	for i := range h.hostOptions {
		h.hostOptions[i].setWidth(width - 2)
	}
	h.ensureSelectionVisible()
}

func (h *HostsInfoModel) setFocus(focused bool) {
	h.focused = focused
	if !focused {
		h.blurCurrentInput()
	}
}

func (h *HostsInfoModel) loadHost(host sqlite.Host) {
	if h.mode == infoEditMode {
		h.ExitEditMode()
	}
	h.host = host.Host
	h.currentEditHost = host
	h.hostNotes.SetValue(host.Notes)
	h.tagsInput.SetValue(strings.Join(host.Tags, ","))
	h.HostPreviewString = buildHostPreview(host)
	h.setHostOptions(host)
	h.selected = h.firstEditableIndex()
	if h.selected < 0 {
		h.selected = h.tagSelectionIndex()
	}
	h.ensureSelectionVisible()
}

func (h *HostsInfoModel) clearHost() {
	h.host = ""
	h.currentEditHost = sqlite.Host{}
	h.hostOptions = nil
	h.HostPreviewString = ""
	h.hostNotes.SetValue("")
	h.tagsInput.SetValue("")
	h.selected = 0
	h.pendingSave = false
}

func (h *HostsInfoModel) EnterEditMode() tea.Cmd {
	if h.host == "" {
		return nil
	}
	h.mode = infoEditMode
	if !h.isIndexEditable(h.selected) {
		h.selected = h.firstEditableIndex()
	}
	if h.selected < 0 {
		h.selected = h.tagSelectionIndex()
	}
	h.ensureSelectionVisible()
	return h.focusCurrentInput()
}

func (h *HostsInfoModel) ExitEditMode() {
	if h.mode != infoEditMode {
		return
	}
	h.blurCurrentInput()
	h.mode = infoViewMode
}

func (h *HostsInfoModel) moveSelection(delta int) tea.Cmd {
	maxIndex := h.totalSelections() - 1
	next := h.selected + delta
	if next < 0 {
		next = 0
	}
	if next > maxIndex {
		next = maxIndex
	}
	if h.mode == infoEditMode {
		step := 1
		if delta < 0 {
			step = -1
		}
		for next >= 0 && next <= maxIndex && !h.isIndexEditable(next) {
			next += step
		}
		if next < 0 {
			next = 0
		}
		if next > maxIndex {
			next = maxIndex
		}
	}
	if next == h.selected {
		return nil
	}
	h.blurCurrentInput()
	h.selected = next
	h.ensureSelectionVisible()
	if h.mode == infoEditMode {
		return h.focusCurrentInput()
	}
	return nil
}

func (h *HostsInfoModel) ensureSelectionVisible() {
	if len(h.hostOptions) == 0 {
		h.optionsScrollPane.YOffset = 0
		return
	}
	if h.selected >= len(h.hostOptions) {
		return
	}
	if h.selected < h.optionsScrollPane.YOffset {
		h.optionsScrollPane.YOffset = h.selected
	} else if h.selected >= h.optionsScrollPane.YOffset+h.optionsScrollPane.Height {
		h.optionsScrollPane.YOffset = h.selected - h.optionsScrollPane.Height + 1
	}
	if h.optionsScrollPane.YOffset < 0 {
		h.optionsScrollPane.YOffset = 0
	}
}

func (h HostsInfoModel) selectionIsNotes() bool {
	return h.selected == h.notesSelectionIndex()
}

func (h HostsInfoModel) selectionIsTags() bool {
	return h.selected == h.tagSelectionIndex()
}

func (h HostsInfoModel) isIndexEditable(idx int) bool {
	if idx == h.tagSelectionIndex() || idx == h.notesSelectionIndex() {
		return true
	}
	if idx < 0 || idx >= len(h.hostOptions) {
		return false
	}
	return !h.hostOptions[idx].neverEditable
}

func (h HostsInfoModel) tagSelectionIndex() int {
	return len(h.hostOptions)
}

func (h HostsInfoModel) notesSelectionIndex() int {
	return len(h.hostOptions) + 1
}

func (h HostsInfoModel) totalSelections() int {
	return len(h.hostOptions) + 2
}

func (h *HostsInfoModel) focusCurrentInput() tea.Cmd {
	if h.mode != infoEditMode {
		return nil
	}
	if h.selectionIsTags() {
		return h.tagsInput.Focus()
	}
	if h.selectionIsNotes() {
		return h.hostNotes.Focus()
	}
	if h.selected < 0 || h.selected >= len(h.hostOptions) {
		return nil
	}
	if h.hostOptions[h.selected].neverEditable {
		return nil
	}
	return h.hostOptions[h.selected].val.Focus()
}

func (h *HostsInfoModel) blurCurrentInput() {
	if h.selectionIsTags() {
		h.tagsInput.Blur()
		return
	}
	if h.selectionIsNotes() {
		h.hostNotes.Blur()
		return
	}
	if h.selected < 0 || h.selected >= len(h.hostOptions) {
		return
	}
	h.hostOptions[h.selected].val.Blur()
}

func (h *HostsInfoModel) setHostOptions(host sqlite.Host) {
	options := make([]kvInputModel, 0, len(host.Options)+1)
	hostRow := newKvInputModel("Host", host.Host, false, true)
	hostRow.setWidth(max(10, h.width-2))
	hostRow.key.Blur()
	hostRow.val.Blur()
	options = append(options, hostRow)

	opts := slices.Clone(host.Options)
	sort.Slice(opts, func(i, j int) bool {
		return opts[i].Key < opts[j].Key
	})
	for _, opt := range opts {
		row := newKvInputModel(opt.Key, opt.Value, false, false)
		row.setWidth(max(10, h.width-2))
		row.key.Blur()
		row.val.Blur()
		options = append(options, row)
	}
	h.hostOptions = options
}

func (h HostsInfoModel) firstEditableIndex() int {
	for idx := range h.hostOptions {
		if !h.hostOptions[idx].neverEditable {
			return idx
		}
	}
	return h.tagSelectionIndex()
}

func (h *HostsInfoModel) buildUpdatedHost() sqlite.Host {
	host := h.currentEditHost
	host.Notes = h.hostNotes.Value()
	host.Tags = parseTagsInput(h.tagsInput.Value())
	host.Options = make([]sqlite.HostOptions, 0, len(h.hostOptions))
	for _, opt := range h.hostOptions {
		if strings.EqualFold(opt.key.Value(), "Host") {
			continue
		}
		host.Options = append(host.Options, sqlite.HostOptions{
			Key:   opt.key.Value(),
			Value: opt.val.Value(),
			Host:  host.Host,
		})
	}
	now := time.Now()
	host.UpdatedAt = &now
	return host
}

func (h HostsInfoModel) renderOptions() string {
	if len(h.hostOptions) == 0 {
		return "No options configured"
	}
	lines := make([]string, len(h.hostOptions))
	for idx, opt := range h.hostOptions {
		indicator := "  "
		if h.focused && h.selected == idx {
			if h.mode == infoEditMode {
				indicator = "> "
			} else {
				indicator = "* "
			}
		}
		if h.mode == infoEditMode && h.selected == idx && !opt.neverEditable {
			lines[idx] = fmt.Sprintf("%s%s %s", indicator, opt.key.Value(), opt.val.View())
		} else {
			lines[idx] = fmt.Sprintf("%s%s: %s", indicator, opt.key.Value(), opt.val.Value())
		}
		if opt.neverEditable {
			lines[idx] += " (locked)"
		}
	}
	return strings.Join(lines, "\n")
}

func (h HostsInfoModel) renderNotes() string {
	if h.mode == infoEditMode && h.selectionIsNotes() {
		return h.hostNotes.View()
	}
	value := strings.TrimSpace(h.hostNotes.Value())
	if value == "" {
		return "(no notes)"
	}
	return value
}

func (h HostsInfoModel) renderTags() string {
	if h.mode == infoEditMode && h.selectionIsTags() {
		return h.tagsInput.View()
	}
	value := strings.TrimSpace(h.tagsInput.Value())
	if value == "" {
		return "(no tags)"
	}
	return value
}

func (h HostsInfoModel) shouldRenderPreview() bool {
	if h.previewCollapsed {
		return false
	}
	if h.height < 20 {
		return false
	}
	return strings.TrimSpace(h.HostPreviewString) != ""
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
	width, height   int
	verticalLayout  bool
}

func NewHostsPanelModel(cfg config.Config, hosts []sqlite.Host) HostsPanelModel {
	tableModel := NewHostsModel(cfg)
	infoModel := NewHostsInfoModel()
	panel := HostsPanelModel{
		table:           tableModel,
		infoPanel:       infoModel,
		focus:           focusTable,
		data:            make([]sqlite.Host, len(hosts)),
		tableGrowthBias: defaultTableBias,
	}
	copy(panel.data, hosts)
	panel.table.setFocused(true)
	panel.refreshTableRows()
	panel.syncInfoWithSelection()
	return panel
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
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.applySize(msg.Width, msg.Height)
	case newHostsMessage:
		h = h.upsertHost(msg.host)
	case updateHostsMessage:
		h = h.upsertHost(msg.host)
	}

	h.table.setFocused(h.focus == focusTable)
	h.infoPanel.setFocus(h.focus == focusInfoPanel)

	tableModel, tableCmd := h.table.Update(msg)
	if tableCmd != nil {
		cmds = append(cmds, tableCmd)
	}
	h.table = tableModel.(HostsModel)

	infoModel, infoCmd := h.infoPanel.Update(msg)
	if infoCmd != nil {
		cmds = append(cmds, infoCmd)
	}
	h.infoPanel = infoModel.(HostsInfoModel)

	if h.infoPanel.pendingSave {
		h = h.upsertHost(h.infoPanel.currentEditHost)
		h.infoPanel.pendingSave = false
	}

	h.syncInfoWithSelection()

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if h.focus == focusTable {
			switch {
			case key.Matches(keyMsg, tableKeyMap.Edit):
				if cmd := h.beginEditSelectedHost(); cmd != nil {
					h.focus = focusInfoPanel
					h.table.setFocused(false)
					cmds = append(cmds, cmd)
				}
			case key.Matches(keyMsg, tableKeyMap.CycleView):
				h.focus = focusInfoPanel
				h.table.setFocused(false)
			case key.Matches(keyMsg, tableKeyMap.Select):
				if host := h.table.highlightedHost(); host != nil {
					cmds = append(cmds, startConnectCmd(*host))
				}
			}
		} else {
			switch {
			case key.Matches(keyMsg, infoPanelKeyMap.ChangeView), key.Matches(keyMsg, infoPanelKeyMap.CancelView):
				h.focus = focusTable
				h.table.setFocused(true)
				h.infoPanel.ExitEditMode()
			}
		}
	}

	return h, tea.Batch(cmds...)
}

func (h HostsPanelModel) View() string {
	tableView := lipgloss.NewStyle().
		Width(h.table.width).
		MarginRight(1).
		Render(h.table.View())
	infoView := lipgloss.NewStyle().
		Width(h.infoPanel.width).
		Render(h.infoPanel.View())
	if h.verticalLayout {
		return lipgloss.JoinVertical(lipgloss.Left, tableView, infoView)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tableView, infoView)
}

func (h *HostsPanelModel) applySize(width, height int) {
	h.width = width
	h.height = height
	h.verticalLayout = width < verticalLayoutThreshold
	if h.verticalLayout {
		tableHeight := height / 2
		infoHeight := height - tableHeight
		h.table.setSize(width, tableHeight)
		h.infoPanel.setSize(width, infoHeight)
		return
	}
	tableWidth := int(float32(width) * h.tableGrowthBias)
	if tableWidth < minimumTableWidth {
		tableWidth = minimumTableWidth
	}
	infoWidth := width - tableWidth
	if infoWidth < minimumInfoWidth {
		infoWidth = minimumInfoWidth
		tableWidth = width - infoWidth
	}
	h.table.setSize(tableWidth, height)
	h.infoPanel.setSize(infoWidth, height)
}

func (h *HostsPanelModel) beginEditSelectedHost() tea.Cmd {
	host := h.table.highlightedHost()
	if host == nil {
		return nil
	}
	h.infoPanel.loadHost(*host)
	return h.infoPanel.EnterEditMode()
}

func (h *HostsPanelModel) refreshTableRows() {
	rows := make([]table.Row, 0, len(h.data))
	for i := range h.data {
		rows = append(rows, hostToRow(&h.data[i], h.table.cfg))
	}
	h.table.setRows(rows)
}

func (h *HostsPanelModel) syncInfoWithSelection() {
	host := h.table.highlightedHost()
	if host == nil {
		h.infoPanel.clearHost()
		return
	}
	if host.Host != h.infoPanel.currentEditHost.Host {
		h.infoPanel.loadHost(*host)
	}
}

func (h HostsPanelModel) upsertHost(host sqlite.Host) HostsPanelModel {
	idx := slices.IndexFunc(h.data, func(existing sqlite.Host) bool {
		return existing.Host == host.Host
	})
	if idx >= 0 {
		h.data[idx] = host
	} else {
		h.data = append(h.data, host)
	}
	h.refreshTableRows()
	h.syncInfoWithSelection()
	return h
}

type connectHostMessage struct {
	host sqlite.Host
}

func hostToRow(host *sqlite.Host, cfg config.Config) table.Row {
	row := table.NewRow(table.RowData{
		hostColumnKey:              host.Host,
		hostHostnameColumnKey:      hostOptionValue(host, "Hostname"),
		hostLastConnectedColumnKey: formatLastConnected(host.LastConnection),
		hostPingColumnKey:          formatPing(cfg.EnablePing),
		hostStatusColumnKey:        formatHostStatus(cfg.EnablePing),
		hostRowPayloadKey:          host,
	})
	return row
}

func hostOptionValue(host *sqlite.Host, key string) string {
	for _, opt := range host.Options {
		if strings.EqualFold(opt.Key, key) {
			return opt.Value
		}
	}
	return ""
}

func formatLastConnected(ts *time.Time) string {
	if ts == nil {
		return "never"
	}
	return ts.Format("2006-01-02 15:04")
}

func formatPing(enabled bool) string {
	if !enabled {
		return "disabled"
	}
	return "n/a"
}

func formatHostStatus(pingEnabled bool) string {
	if !pingEnabled {
		return "unknown"
	}
	return "pending"
}

func startConnectCmd(host sqlite.Host) tea.Cmd {
	return func() tea.Msg {
		return connectHostMessage{host: host}
	}
}

func buildHostPreview(host sqlite.Host) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Host %s\n", host.Host))
	opts := slices.Clone(host.Options)
	sort.Slice(opts, func(i, j int) bool {
		if opts[i].Key == opts[j].Key {
			return opts[i].Value < opts[j].Value
		}
		return opts[i].Key < opts[j].Key
	})
	for _, opt := range opts {
		builder.WriteString(fmt.Sprintf("  %s %s\n", opt.Key, opt.Value))
	}
	if len(host.Tags) > 0 {
		builder.WriteString("Tags: ")
		builder.WriteString(strings.Join(host.Tags, ","))
		builder.WriteRune('\n')
	}
	if note := strings.TrimSpace(host.Notes); note != "" {
		builder.WriteString("# ")
		builder.WriteString(note)
		builder.WriteRune('\n')
	}
	return strings.TrimRight(builder.String(), "\n")
}

func parseTagsInput(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	tags := make([]string, 0, len(parts))
	seen := make(map[string]struct{})
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}
	return tags
}
