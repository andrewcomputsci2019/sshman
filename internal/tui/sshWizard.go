package tui

import (
	"andrew/sshman/internal/sshUtils"
	"errors"
	"slices"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	keyInputFocusState = iota
	valueInputFocusState
)

const (
	formNavigateMode = iota
	formEditMode
)

// todo implement kvRowInput Form logic
type kvRowInput struct {
	key        textinput.Model
	val        textinput.Model
	focus      bool // weather the row has focus
	mode       int  // navigation or edit
	inputFocus int
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
				if k.inputFocus == 0 {
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
				if k.inputFocus == 0 {
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
			if k.inputFocus == 0 {
				k.inputFocus = 1
			}
			return k, nil
		}
		if (msg.Type == tea.KeyShiftTab || msg.Type == tea.KeyLeft) && k.mode == formNavigateMode {
			if k.inputFocus == 1 {
				k.inputFocus = 0
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
	//TODO implement me
	panic("implement me")
}

type WizardViewModel struct {
	hostInput     textinput.Model
	hostnameInput textinput.Model
	hostOptions   []kvRowInput
	notes         textarea.Model
	selectedRow   int // 0 hostInput, 1 hostnameInput, 2 from []kvRowInput onwards, len(hostOptions)+2 textarea , and len(hostOptions)+3 == confirm button
	mode          int // edit means a input has focus
	width, height int
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
			return w, nil
		}
		if msg.String() == "up" || msg.String() == "j" || msg.Type == tea.KeyShiftTab {
			if w.selectedRow == 0 {
				return w, nil
			}
			w.selectedRow--
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
				w.hostOptions = append(w.hostOptions, kvRowInput{}) // as a user adds entries we
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
					w.hostOptions[index] = kvRowInput{}
				}
				return w, nil
			}
		}
		return w, nil
	default:
		return w, nil
	}
}

func (w WizardViewModel) View() string {
	//TODO implement me
	panic("implement me")
}

func NewWizardViewModel() WizardViewModel {
	hostInput := textinput.New()
	hostnameInput := textinput.New()
	hostInput.Prompt = "Host (match rule)/alias"
	hostnameInput.Prompt = "Hostname option"
	hostnameInput.Validate = hostValidatorWrapper
	return WizardViewModel{
		hostInput:     textinput.New(),
		hostnameInput: textinput.New(),
		hostOptions:   make([]kvRowInput, 2),
		notes:         textarea.New(),
		selectedRow:   0,
		mode:          formNavigateMode,
		width:         0,
		height:        0,
	}
}

func hostValidatorWrapper(h string) error {
	if !sshUtils.ValidHost(h) {
		return errors.New("invalid host")
	}
	return nil
}
