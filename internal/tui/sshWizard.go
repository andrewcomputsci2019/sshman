package tui

import (
	"andrew/sshman/internal/sqlite"
	"andrew/sshman/internal/sshUtils"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	keyInputFocusState = iota
	valueInputFocusState
)

const (
	formNavigateMode = iota
	formEditMode
)

const (
	wizardDefaultFormWidth   = 60
	wizardMaxFormWidth       = 100
	wizardMinFormWidth       = 40
	wizardNotesHeight        = 6
	wizardHostFieldHeight    = 3 // 3 single-line host inputs stacked
	wizardConfirmationHeight = 1
	wizardViewportMinRows    = 3
	wizardViewportMaxRows    = 5
	wizardViewportMinHeight  = wizardViewportMinRows * kvRowHeight
	wizardViewportMaxHeight  = wizardViewportMaxRows * kvRowHeight
	kvRowHeight              = 3
	kvRowSpacerWidth         = 0
	kvRowHorizontalPadding   = 1
	kvRowMinInputWidth       = 8
	wizardIndicatorWidth     = 2
)

var selectionIndicatorStyle = lipgloss.NewStyle().Width(wizardIndicatorWidth)

func selectionIndicator(selected bool) string {
	if selected {
		return selectionIndicatorStyle.Render("> ")
	}
	return selectionIndicatorStyle.Render("  ")
}

// todo implement kvRowInput Form logic
type kvRowInput struct {
	key        textinput.Model
	val        textinput.Model
	focus      bool // weather the row has focus
	mode       int  // navigation or edit
	inputFocus int
	width      int
	height     int
}

func newKVRowInput() kvRowInput {
	key := textinput.New()
	key.Placeholder = "Option key"
	val := textinput.New()
	val.Placeholder = "Option value"
	key.SetSuggestions(sshUtils.GetListOfAcceptableOptions())
	key.ShowSuggestions = true
	key.CompletionStyle = key.CompletionStyle.Foreground(lipgloss.Color("255"))
	key.TextStyle = key.TextStyle.Foreground(lipgloss.Color("#6CB6EB"))
	val.CompletionStyle = val.CompletionStyle.Foreground(lipgloss.Color("255"))
	val.TextStyle = val.TextStyle.Foreground(lipgloss.Color("#E5E7EB"))
	return kvRowInput{
		key:    key,
		val:    val,
		height: kvRowHeight,
	}
}

func (k *kvRowInput) SetWidth(total int) {
	if total <= 0 {
		return
	}
	k.width = total - 2
	// so its - (2*2) for prompt width, -2 for border
	innerWidth := max(total-8, kvRowMinInputWidth*2+kvRowSpacerWidth)
	keyWidth := (innerWidth) / 2
	valWidth := innerWidth - keyWidth
	if keyWidth < kvRowMinInputWidth {
		keyWidth = kvRowMinInputWidth
	}
	if valWidth < kvRowMinInputWidth {
		valWidth = kvRowMinInputWidth
	}
	k.key.Width = keyWidth
	k.val.Width = valWidth
}

func updateKVModel(k kvRowInput, msg tea.Msg) (kvRowInput, tea.Cmd) {
	update, cmd := k.Update(msg)
	cast := update.(kvRowInput)
	return cast, cmd
}

func (k kvRowInput) Init() tea.Cmd {
	// no-op
	return nil
}

func setPromptStyleKvRowInput(input *textinput.Model, focused bool) {
	if focused {
		input.PromptStyle = input.PromptStyle.Foreground(lipgloss.Color("#7D56F4"))
	} else {
		input.PromptStyle = input.PromptStyle.Foreground(lipgloss.Color("255"))
	}
}

func (k kvRowInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEscape {
			if k.mode == formNavigateMode {
				k.focus = false
				setPromptStyleKvRowInput(&k.key, false)
				setPromptStyleKvRowInput(&k.val, false)
				return k, nil
			} else {
				k.mode = formNavigateMode
				if k.inputFocus == keyInputFocusState {
					k.key.SetValue("")
					k.key.Blur()
				} else {
					k.val.SetValue("")
					k.val.Blur()
				}
				return k, nil
			}
		}
		if msg.Type == tea.KeyEnter {
			if k.mode == formEditMode {
				k.mode = formNavigateMode
				if k.inputFocus == keyInputFocusState {
					k.key.Blur()
					if sshUtils.IsAcceptableOption(k.key.Value()) {
						if sshUtils.IsOptionYesNo(k.key.Value()) {
							k.val.SetSuggestions([]string{"yes", "no"})
							k.val.ShowSuggestions = true
						} else if k.key.Value() == "AddressFamily" {
							k.val.SetSuggestions(sshUtils.GetAllAddressFamily())
						}
					}
				} else {
					k.val.ShowSuggestions = false
					k.val.Blur()
				}
				return k, nil
			} else {
				k.mode = formEditMode
				if k.inputFocus == 0 {
					cmd := k.key.Focus()
					return k, cmd
				}
				cmd := k.val.Focus()
				return k, cmd
			}
		}
		if (msg.Type == tea.KeyTab || msg.Type == tea.KeyRight) && k.mode == formNavigateMode {
			if k.inputFocus == keyInputFocusState {
				k.inputFocus = valueInputFocusState
				setPromptStyleKvRowInput(&k.key, false)
				setPromptStyleKvRowInput(&k.val, true)
			}
			return k, nil
		}
		if (msg.Type == tea.KeyShiftTab || msg.Type == tea.KeyLeft) && k.mode == formNavigateMode {
			if k.inputFocus == valueInputFocusState {
				k.inputFocus = keyInputFocusState
				setPromptStyleKvRowInput(&k.key, true)
				setPromptStyleKvRowInput(&k.val, false)
			}
			return k, nil
		}
		if k.mode == formEditMode {
			if k.inputFocus == 0 {
				input, cmd := k.key.Update(msg)
				k.key = input
				return k, cmd
			}
			input, cmd := k.val.Update(msg)
			k.val = input
			return k, cmd
		}
		return k, nil // in navigation mode user types a key that is not known skip
	default:
		return k, nil
	}
}

func (k kvRowInput) View() string {
	keyView := k.key.View()
	valView := k.val.View()

	content := lipgloss.JoinHorizontal(lipgloss.Left, keyView, valView)

	borderColor := lipgloss.Color("255")
	if k.focus {
		borderColor = lipgloss.Color("#7D56F4")
	}

	rowStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(0).
		Margin(0)

	if k.width > 0 {
		rowStyle = rowStyle.Width(k.width)
	}

	return rowStyle.Render(content)
}

type WizardViewModel struct {
	hostInput     textinput.Model
	hostnameInput textinput.Model
	tags          textinput.Model
	hostOptions   []kvRowInput
	notes         textarea.Model
	selectedRow   int // 0 hostInput, 1 hostnameInput, 2 from []kvRowInput onwards, len(hostOptions)+2 textarea , and len(hostOptions)+3 == confirm button
	mode          int // edit means a input has focus
	kvViewport    viewport.Model
	formWidth     int
	width, height int
}

func (w WizardViewModel) innerWidth() int {
	width := w.formWidth
	if width == 0 {
		width = wizardDefaultFormWidth
	}
	width -= wizardIndicatorWidth
	if width < 1 {
		width = 1
	}
	return width
}

func (w *WizardViewModel) recalcLayout() {
	formWidth := w.width
	if formWidth == 0 {
		formWidth = wizardDefaultFormWidth
	}
	if formWidth > wizardMaxFormWidth {
		formWidth = wizardMaxFormWidth
	}
	if formWidth < wizardMinFormWidth {
		formWidth = wizardMinFormWidth
	}
	w.formWidth = formWidth
	contentWidth := w.innerWidth()

	w.hostInput.Width = contentWidth
	w.hostnameInput.Width = contentWidth
	w.tags.Width = contentWidth
	w.notes.SetWidth(contentWidth)
	w.notes.SetHeight(wizardNotesHeight)

	for i := range w.hostOptions {
		w.hostOptions[i].SetWidth(contentWidth)
	}

	w.kvViewport.Width = formWidth
	fixedVertical := w.notes.Height() + wizardHostFieldHeight + wizardConfirmationHeight
	availableHeight := w.height - fixedVertical
	switch {
	case availableHeight <= 0:
		w.kvViewport.Height = kvRowHeight
	case availableHeight < wizardViewportMinHeight:
		w.kvViewport.Height = availableHeight
	case availableHeight > wizardViewportMaxHeight:
		w.kvViewport.Height = wizardViewportMaxHeight
	default:
		w.kvViewport.Height = availableHeight
	}
	w.ensureKVSelectionVisible()
}

func (w *WizardViewModel) ensureKVSelectionVisible() {
	if len(w.hostOptions) == 0 {
		w.kvViewport.SetYOffset(0)
		return
	}
	index := w.selectedRow - 2
	if index < 0 {
		w.kvViewport.SetYOffset(0)
		return
	}
	if index >= len(w.hostOptions) {
		index = len(w.hostOptions) - 1
	}
	rowStart := index * kvRowHeight
	rowEnd := rowStart + kvRowHeight
	top := w.kvViewport.YOffset
	bottom := top + w.kvViewport.Height
	if rowStart < top {
		w.kvViewport.SetYOffset(rowStart)
		return
	}
	if rowEnd > bottom {
		offset := max(rowEnd-w.kvViewport.Height, 0)
		w.kvViewport.SetYOffset(offset)
	}
}

func (w WizardViewModel) Init() tea.Cmd {

	return nil
}

func (w WizardViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if w.mode != formNavigateMode && w.mode != formEditMode {
		w.mode = formNavigateMode
	}
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		w.width = msg.Width
		w.height = msg.Height
		w.recalcLayout()
		return w, nil
	}
	if w.mode == formEditMode {
		if w.selectedRow < 3 {
			switch w.selectedRow {
			case 0:
				if key, ok := msg.(tea.KeyMsg); ok {
					if key.Type == tea.KeyEnter || key.Type == tea.KeyEsc {
						w.mode = formNavigateMode
						w.hostInput.Blur()
						return w, nil
					}
					input, cmd := w.hostInput.Update(msg)
					w.hostInput = input
					return w, cmd
				} else {
					return w, nil
				}
			case 1:
				if key, ok := msg.(tea.KeyMsg); ok {
					if key.Type == tea.KeyEnter || key.Type == tea.KeyEsc {
						w.mode = formNavigateMode
						w.hostnameInput.Blur()
						return w, nil
					}
					input, cmd := w.hostnameInput.Update(msg)
					w.hostnameInput = input
					return w, cmd
				} else {
					return w, nil
				}
			default:
				if key, ok := msg.(tea.KeyMsg); ok {
					if key.Type == tea.KeyEnter || key.Type == tea.KeyEsc {
						w.mode = formNavigateMode
						w.tags.Blur()
						return w, nil
					}
					input, cmd := w.tags.Update(msg)
					w.tags = input
					return w, cmd
				} else {
					return w, nil
				}
			}
		} else if w.selectedRow < len(w.hostOptions)+3 {
			index := w.selectedRow - 3
			if w.hostOptions[index].focus {
				input, cmd := updateKVModel(w.hostOptions[index], msg)
				w.hostOptions[index] = input
				return w, cmd
			} else {
				w.mode = formNavigateMode
			}
		} else {
			if w.selectedRow == len(w.hostOptions)+3 && w.notes.Focused() {
				if msg, ok := msg.(tea.KeyMsg); ok {
					if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlS {
						w.notes.Blur()
						return w, nil
					}
				}
				notes, cmd := w.notes.Update(msg)
				w.notes = notes
				return w, cmd
			} else {
				w.mode = formNavigateMode
			}
		}
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc {
			return w, func() tea.Msg { return userExitWizard{} }
		}
		if w.selectedRow == len(w.hostOptions)+4 && msg.String() == "enter" {
			cmd := func() tea.Msg {
				host := w.hostInput.Value()
				options := make([]sqlite.HostOptions, 0)
				options = append(options, sqlite.HostOptions{
					ID:    0,
					Key:   "Hostname",
					Value: w.hostnameInput.Value(),
					Host:  host,
				})
				for _, optRow := range w.hostOptions {
					options = append(options, sqlite.HostOptions{
						ID:    0,
						Key:   optRow.key.Value(),
						Value: optRow.val.Value(),
						Host:  host,
					})
				}
				tags := strings.Split(w.tags.Value(), ",")
				return newHostsMessage{
					host: sqlite.Host{
						Host:           host,
						CreatedAt:      time.Now(),
						UpdatedAt:      nil,
						LastConnection: nil,
						Notes:          w.notes.Value(),
						Options:        options,
						Tags:           tags,
					},
				}
			}
			return w, cmd
		}
		if msg.Type == tea.KeyEnd {
			w.selectedRow = len(w.hostOptions) + 4
			w.ensureKVSelectionVisible()
			return w, nil
		}
		if msg.Type == tea.KeyHome {
			w.selectedRow = 0
			w.ensureKVSelectionVisible()
			return w, nil
		}
		if msg.String() == "down" || msg.String() == "k" || msg.String() == "tab" {
			if w.selectedRow == len(w.hostOptions)+4 {
				return w, nil
			}
			w.selectedRow++
			w.ensureKVSelectionVisible()
			return w, nil
		}
		if msg.String() == "up" || msg.String() == "j" || msg.Type == tea.KeyShiftTab {
			if w.selectedRow == 0 {
				return w, nil
			}
			w.selectedRow--
			w.ensureKVSelectionVisible()
			return w, nil
		}
		if msg.String() == "enter" {
			w.mode = formEditMode
			// get selected row and then toggle that element as focused
			if w.selectedRow < 3 {
				switch w.selectedRow {
				case 0:
					cmd := w.hostInput.Focus()
					return w, cmd
				case 1:
					cmd := w.hostnameInput.Focus()
					return w, cmd
				default:
					cmd := w.tags.Focus()
					return w, cmd
				}
			}
			index := w.selectedRow - 3
			if index == len(w.hostOptions)-1 {
				newRow := newKVRowInput()
				newRow.SetWidth(w.innerWidth())
				w.hostOptions = append(w.hostOptions, newRow) // as a user adds entries we
				// want to keep adding options so they can continue to add more
			}
			if index < len(w.hostOptions) {
				w.hostOptions[index].focus = true
				setPromptStyleKvRowInput(&w.hostOptions[index].key, true)
				return w, nil
			}
			if index == len(w.hostOptions) {
				// this is the text area and it needs focus
				cmd := w.notes.Focus()
				return w, cmd
			}
		}
		if msg.String() == "d" {
			index := w.selectedRow - 3
			if index > 0 && index < len(w.hostOptions) {
				// delete that option
				if len(w.hostOptions) > 1 {
					w.hostOptions = slices.Delete(w.hostOptions, index, index+1)
					if index == len(w.hostOptions) {
						w.selectedRow--
					}
				} else { // clear the option if 2 or less rows exist
					w.hostOptions[index] = newKVRowInput()
					w.hostOptions[index].SetWidth(w.innerWidth())
				}
				w.ensureKVSelectionVisible()
				return w, nil
			}
		}
		return w, nil
	default:
		return w, nil
	}
}

func (w WizardViewModel) View() string {
	contentWidth := w.innerWidth()

	rowViews := make([]string, len(w.hostOptions))
	for i := range w.hostOptions {
		row := w.hostOptions[i].View()
		rowViews[i] = lipgloss.JoinHorizontal(
			lipgloss.Left,
			selectionIndicator(w.selectedRow == i+3),
			row,
		)
	}

	if len(rowViews) == 0 {
		rowViews = append(rowViews, lipgloss.JoinHorizontal(lipgloss.Left, selectionIndicator(false), ""))
	}

	w.kvViewport.SetContent(strings.Join(rowViews, "\n"))
	w.ensureKVSelectionVisible()

	formStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Left)
	host := lipgloss.JoinHorizontal(
		lipgloss.Left,
		selectionIndicator(w.selectedRow == 0),
		formStyle.Render(w.hostInput.View()),
	)
	hostname := lipgloss.JoinHorizontal(
		lipgloss.Left,
		selectionIndicator(w.selectedRow == 1),
		formStyle.Render(w.hostnameInput.View()),
	)
	tags := lipgloss.JoinHorizontal(lipgloss.Left,
		selectionIndicator(w.selectedRow == 2),
		formStyle.Render(w.tags.View()),
	)
	notesView := w.notes.View()
	if w.selectedRow == len(w.hostOptions)+3 {
		notesView = lipgloss.NewStyle().
			Background(lipgloss.Color("#2F2F6B")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Render(notesView)
	}
	notes := lipgloss.JoinHorizontal(
		lipgloss.Left,
		selectionIndicator(w.selectedRow == len(w.hostOptions)+3),
		formStyle.Render(notesView),
	)

	confirmStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Padding(0, 1).
		Foreground(lipgloss.Color("#E4E4E7")).
		Background(lipgloss.Color("#2F2F6B"))
	confirmFocusedStyle := confirmStyle
	confirmFocusedStyle = confirmFocusedStyle.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#3D5AFE")).
		Bold(true)
	var confirm string
	if w.selectedRow == len(w.hostOptions)+4 {
		confirm = confirmFocusedStyle.Render("> Confirm <")
	} else {
		confirm = confirmStyle.Render("Confirm")
	}
	confirm = lipgloss.JoinHorizontal(
		lipgloss.Left,
		selectionIndicator(w.selectedRow == len(w.hostOptions)+4),
		confirm,
	)

	form := []string{host, hostname, tags, w.kvViewport.View(), notes, confirm}
	return lipgloss.JoinVertical(lipgloss.Left, form...)
}

func NewWizardViewModel() WizardViewModel {
	hostInput := textinput.New()
	hostInput.Prompt = "Host (match rule)/alias "
	hostInput.Placeholder = "alias"
	hostInput.TextStyle = hostInput.TextStyle.Foreground(lipgloss.Color("#7AA2F7")).Bold(true)

	hostnameInput := textinput.New()
	hostnameInput.Prompt = "Hostname "
	hostnameInput.Placeholder = "example.com"
	hostnameInput.Validate = hostValidatorWrapper
	hostnameInput.TextStyle = hostnameInput.TextStyle.Foreground(lipgloss.Color("#7AA2F7")).Bold(true)

	tagsInput := textinput.New()
	tagsInput.Prompt = "Tags "
	tagsInput.Placeholder = "tag1,tag2"
	tagsInput.TextStyle = hostInput.TextStyle.Foreground(lipgloss.Color("#7AA2F7")).Bold(true)

	notes := textarea.New()
	notes.Placeholder = "Notes"
	notes.SetHeight(wizardNotesHeight)

	hostOptions := make([]kvRowInput, 2)
	for i := range hostOptions {
		hostOptions[i] = newKVRowInput()
	}

	formWidth := wizardDefaultFormWidth
	defaultHeight := wizardHostFieldHeight + wizardConfirmationHeight + wizardNotesHeight + wizardViewportMinHeight
	kvViewPort := viewport.New(formWidth, wizardViewportMinHeight)
	wiz := WizardViewModel{
		hostInput:     hostInput,
		hostnameInput: hostnameInput,
		tags:          tagsInput,
		hostOptions:   hostOptions,
		notes:         notes,
		selectedRow:   0,
		mode:          formNavigateMode,
		kvViewport:    kvViewPort,
		formWidth:     formWidth,
		width:         formWidth,
		height:        defaultHeight,
	}
	wiz.recalcLayout()
	return wiz
}

func hostValidatorWrapper(h string) error {
	if !sshUtils.ValidHost(h) {
		return errors.New("invalid host")
	}
	return nil
}
