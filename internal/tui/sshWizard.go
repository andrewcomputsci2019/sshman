package tui

import (
	"andrew/sshman/internal/sshUtils"
	"errors"
	"slices"
	"strings"

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
	wizardMaxFormWidth       = 72
	wizardMinFormWidth       = 40
	wizardNotesHeight        = 6
	wizardHostFieldHeight    = 2 // two single-line host inputs stacked
	wizardConfirmationHeight = 1
	wizardViewportMinRows    = 3
	wizardViewportMaxRows    = 5
	wizardViewportMinHeight  = wizardViewportMinRows * kvRowHeight
	wizardViewportMaxHeight  = wizardViewportMaxRows * kvRowHeight
	kvRowHeight              = 3
	kvRowSpacerWidth         = 3
	kvRowHorizontalPadding   = 1
	kvRowMinInputWidth       = 8
)

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
	k.width = total
	chromeWidth := 2 + (kvRowHorizontalPadding * 2) // 2 for border
	innerWidth := total - chromeWidth
	if innerWidth < kvRowMinInputWidth*2+kvRowSpacerWidth {
		innerWidth = kvRowMinInputWidth*2 + kvRowSpacerWidth
	}
	keyWidth := (innerWidth - kvRowSpacerWidth) / 2
	valWidth := innerWidth - kvRowSpacerWidth - keyWidth
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
	return nil
}

func (k kvRowInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEscape {
			if k.mode == formNavigateMode {
				k.focus = false
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
			// handle keystroke here
			if k.mode == formEditMode {
				k.mode = formNavigateMode
				if k.inputFocus == keyInputFocusState {
					k.key.Blur()
				} else {
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
			}
			return k, nil
		}
		if (msg.Type == tea.KeyShiftTab || msg.Type == tea.KeyLeft) && k.mode == formNavigateMode {
			if k.inputFocus == valueInputFocusState {
				k.inputFocus = keyInputFocusState
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

	spacer := lipgloss.NewStyle().Width(kvRowSpacerWidth).Render(" ")
	content := lipgloss.JoinHorizontal(lipgloss.Left, keyView, spacer, valView)

	borderColor := lipgloss.Color("#5A5A5A")
	if k.focus {
		borderColor = lipgloss.Color("#7D56F4")
	}

	rowStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	if k.width > 0 {
		rowStyle = rowStyle.Width(k.width)
	}

	return rowStyle.Render(content)
}

type WizardViewModel struct {
	hostInput     textinput.Model
	hostnameInput textinput.Model
	hostOptions   []kvRowInput
	notes         textarea.Model
	selectedRow   int // 0 hostInput, 1 hostnameInput, 2 from []kvRowInput onwards, len(hostOptions)+2 textarea , and len(hostOptions)+3 == confirm button
	mode          int // edit means a input has focus
	kvViewport    viewport.Model
	formWidth     int
	width, height int
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

	w.hostInput.Width = formWidth
	w.hostnameInput.Width = formWidth
	w.notes.SetWidth(formWidth)
	w.notes.SetHeight(wizardNotesHeight)

	for i := range w.hostOptions {
		w.hostOptions[i].SetWidth(formWidth)
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
	if w.mode != formNavigateMode || w.mode != formEditMode {
		w.mode = formNavigateMode
	}
	if w.mode == formEditMode {
		if w.selectedRow < 2 {
			if w.selectedRow == 0 {
				if w.hostInput.Focused() {
					input, cmd := w.hostInput.Update(msg)
					w.hostInput = input
					return w, cmd
				} else {
					w.mode = formNavigateMode
				}
			} else {
				if w.hostnameInput.Focused() {
					input, cmd := w.hostnameInput.Update(msg)
					w.hostnameInput = input
					return w, cmd
				} else {
					w.mode = formNavigateMode
				}
			}
		} else if w.selectedRow < len(w.hostOptions)+2 {
			index := w.selectedRow - 2
			if w.hostOptions[index].focus {
				input, cmd := updateKVModel(w.hostOptions[index], msg)
				w.hostOptions[index] = input
				return w, cmd
			} else {
				w.mode = formNavigateMode
			}
		} else {
			if w.selectedRow == len(w.hostOptions)+2 && w.notes.Focused() {
				notes, cmd := w.notes.Update(msg)
				w.notes = notes
				return w, cmd
			} else {
				w.mode = formNavigateMode
			}
		}
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
		w.recalcLayout()
		return w, nil
	case tea.KeyMsg:
		if w.selectedRow == len(w.hostOptions)+3 && msg.String() == "enter" {
			// todo this is what confirms the form
			return w, nil
		}
		if msg.String() == "down" || msg.String() == "k" || msg.String() == "tab" {
			if w.selectedRow == len(w.hostOptions)+2 {
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
			if w.selectedRow < 2 {
				if w.selectedRow == 0 {
					cmd := w.hostInput.Focus()
					return w, cmd
				}
				cmd := w.hostnameInput.Focus()
				return w, cmd
			}
			index := w.selectedRow - 2
			if index == len(w.hostOptions)-1 {
				newRow := newKVRowInput()
				newRow.SetWidth(w.formWidth)
				w.hostOptions = append(w.hostOptions, newRow) // as a user adds entries we
				// want to keep adding options so they can continue to add more
			}
			if index < len(w.hostOptions) {
				w.hostOptions[index].focus = true
				return w, nil
			}
			if index == len(w.hostOptions) {
				// this is the text area and it needs focus
				cmd := w.notes.Focus()
				return w, cmd
			}
		}
		if msg.String() == "d" {
			index := w.selectedRow - 2
			if index > 2 && index < len(w.hostOptions) {
				// delete that option
				if len(w.hostOptions) > 1 {
					w.hostOptions = slices.Delete(w.hostOptions, index, index+1)
				} else { // clear the option if 2 or less rows exist
					w.hostOptions[index] = newKVRowInput()
					w.hostOptions[index].SetWidth(w.formWidth)
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
	formWidth := w.formWidth
	if formWidth == 0 {
		formWidth = wizardDefaultFormWidth
	}

	rowViews := make([]string, len(w.hostOptions))
	for i := range w.hostOptions {
		rowViews[i] = w.hostOptions[i].View()
	}

	if len(rowViews) == 0 {
		rowViews = append(rowViews, "")
	}

	w.kvViewport.SetContent(strings.Join(rowViews, "\n"))
	w.ensureKVSelectionVisible()

	formStyle := lipgloss.NewStyle().Width(formWidth).Align(lipgloss.Left)
	host := formStyle.Render(w.hostInput.View())
	hostname := formStyle.Render(w.hostnameInput.View())
	notes := formStyle.Render(w.notes.View())

	confirmStyle := lipgloss.NewStyle().
		Width(w.formWidth).
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
	if w.selectedRow == len(w.hostOptions)+3 {
		confirm = confirmFocusedStyle.Render("> Confirm <")
	} else {
		confirm = confirmStyle.Render("Confirm")
	}

	form := []string{host, hostname, w.kvViewport.View(), notes, confirm}
	return lipgloss.JoinVertical(lipgloss.Left, form...)
}

func NewWizardViewModel() WizardViewModel {
	hostInput := textinput.New()
	hostInput.Prompt = "Host (match rule)/alias "
	hostInput.Placeholder = "alias"

	hostnameInput := textinput.New()
	hostnameInput.Prompt = "Hostname option "
	hostnameInput.Placeholder = "example.com"
	hostnameInput.Validate = hostValidatorWrapper

	notes := textarea.New()
	notes.Placeholder = "Notes"
	notes.SetHeight(wizardNotesHeight)

	hostOptions := make([]kvRowInput, 2)
	for i := range hostOptions {
		hostOptions[i] = newKVRowInput()
	}

	formWidth := wizardDefaultFormWidth
	defaultHeight := wizardHostFieldHeight + wizardConfirmationHeight + wizardNotesHeight + wizardViewportMinHeight

	wiz := WizardViewModel{
		hostInput:     hostInput,
		hostnameInput: hostnameInput,
		hostOptions:   hostOptions,
		notes:         notes,
		selectedRow:   0,
		mode:          formNavigateMode,
		kvViewport:    viewport.New(formWidth, wizardViewportMinHeight),
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
