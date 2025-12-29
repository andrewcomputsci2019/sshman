package tui

import (
	"andrew/sshman/internal/config"
	"andrew/sshman/internal/sshUtils"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type KeyGenModel struct {
	form          *huh.Form
	width, height int
	spinner       spinner.Model
}

type KeyRotateModel struct {
	form          *huh.Form
	width, height int
	spinner       spinner.Model
}

type abortedKeyGenForm struct{} // tells parent model that the form has been aborted by the user, so it can close the form

type keyGenResult struct { // this informs the parent model not form that the forms is done and key is generated
	host    string
	keyPair sshUtils.KeyPair
	err     error
}

type abortedRotatedKeyForm struct{}

type keyRotateRequest struct {
	host       string
	oldKeyPath string           // path to the old key to be replaced, if empty user wants to just add new key to server
	newKeySet  sshUtils.KeyPair // new key set
	err        error
}

// form fields keys
const (
	KEY_GEN_ALGO_STR_KEY      = "KEY_GEN_ALGO"
	KEY_GEN_PASSWORD          = "KEY_GEN_PASSWORD"
	KEY_GEN_PASSWORD_VALIDATE = "KEY_GEN_PASSWORD_VALIDATE"
	OLD_KEY_GEN_TO_REPLACE    = "OLD_KEY_REPLACE"
	ROTATE_KEY_STR_KEY        = "SELECTED_KEY_TO_ROTATE"
)

func getTheme() *huh.Theme {
	theme := huh.ThemeBase()
	purple := lipgloss.Color("#7D56F4")
	theme.Focused.Base = theme.Focused.Base.BorderForeground(purple)
	theme.Focused.SelectSelector = theme.Focused.SelectSelector.Foreground(purple)
	theme.Focused.TextInput.Prompt = theme.Focused.TextInput.Prompt.Foreground(purple)
	theme.Focused.TextInput.Cursor = theme.Focused.TextInput.Cursor.Foreground(purple)
	theme.Focused.FocusedButton = theme.Focused.FocusedButton.Background(purple)
	return theme
}

func NewKeyGenModel(host string, cfg config.Config) KeyGenModel {
	// todo create key-gen form
	// the form title should be the host followed by key gen
	// should have just a single option which is a selector of acceptable key gen options from the config file
	// followed by a confirm button
	keyGenOptions := make([]string, 0)
	if len(cfg.Ssh.AcceptableKeyGenAlgorithms) > 0 {
		for _, keyAlg := range cfg.Ssh.AcceptableKeyGenAlgorithms {
			keyGenOptions = append(keyGenOptions, strings.ToUpper(keyAlg))
		}
	} else {
		keyGenOptions = append(keyGenOptions, config.RSA, config.ECDSA, config.ED25519)
	}
	var password string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Key(KEY_GEN_ALGO_STR_KEY).
			Options(huh.NewOptions(keyGenOptions...)...).
			Title("Key Generation Method").Validate(func(s string) error {
			if len(s) == 0 {
				return fmt.Errorf("A option must be selected")
			}
			return nil
		}),
		huh.NewInput().
			Title("Password").
			EchoMode(huh.EchoModePassword).
			Key(KEY_GEN_PASSWORD).Value(&password),
		huh.NewInput().
			Title("Retype Password").
			Key(KEY_GEN_PASSWORD_VALIDATE).
			EchoMode(huh.EchoModePassword).Validate(func(s string) error {
			if len(password) == 0 && len(s) != 0 {
				return fmt.Errorf("Password field is empty")
			}
			if password != s {
				return fmt.Errorf("Passwords do not match")
			}
			return nil
		}),
		huh.NewConfirm().
			Title("Generate Key").
			Affirmative("Yes").
			Negative("No"),
	).
		Title(host + " Key Generation form")).
		WithShowHelp(true).
		WithWidth(45).
		WithShowErrors(true).
		WithTheme(getTheme())
	// todo set form Submit cmd and Cancel cmd to correct types
	form.CancelCmd = func() tea.Msg {
		return abortedKeyGenForm{}
	}
	form.SubmitCmd = func() tea.Msg {
		keyGenType := form.GetString(KEY_GEN_ALGO_STR_KEY)
		password := form.GetString(KEY_GEN_PASSWORD)
		hostString := host
		keyPair, err := sshUtils.GenKey(hostString, keyGenType, password, cfg)
		return keyGenResult{
			host:    hostString,
			err:     err,
			keyPair: keyPair,
		}
	}
	return KeyGenModel{
		width:   45,
		height:  20,
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		form:    form,
	}
}

func (km KeyGenModel) Init() tea.Cmd {
	return km.form.Init()
}

func (km KeyGenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		km.width = msg.Width
		km.height = msg.Height
		form, cmd := km.form.Update(msg)
		if updForm, ok := form.(*huh.Form); ok {
			km.form = updForm
		}
		return km, cmd
	case spinner.TickMsg:
		spin, cmd := km.spinner.Update(msg)
		km.spinner = spin
		return km, cmd
	}
	if km.form.State != huh.StateNormal { // dont send messages to the form if its done processing
		return km, nil
	}
	// todo send msg to form and check if state has changed to abort or submit
	cmds := make([]tea.Cmd, 0)
	form, cmd := km.form.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if updForm, ok := form.(*huh.Form); ok {
		km.form = updForm
	}
	switch km.form.State {
	case huh.StateCompleted:
		cmds = append(cmds, km.spinner.Tick)
	case huh.StateAborted:
		slog.Info("User aborted key generation huh form", "time", time.Now())
	}

	return km, tea.Batch(cmds...)

}

func (km KeyGenModel) View() string {
	switch km.form.State {
	case huh.StateCompleted:
		return fmt.Sprintf("%s Generating keys...", km.spinner.View())
	case huh.StateAborted:
		return ""
	default:
		return km.form.View()
	}
}

func NewKeyRotateModel(host string, keys []string, cfg config.Config) KeyRotateModel {
	// todo create a form have name host followed key rotation, similar to the key gen one except
	// file selector (could use a list here filter beforehand on form creation looking over keystore directory finding valid keys to look for)
	// then again ask for a key gen algorithm from config passed in
	// and then a confirm button
	ownedKeys := make([]string, 0)
	ownedKeys = append(ownedKeys, "None")
	for _, key := range keys {
		if strings.HasPrefix(key, "~/") { // in case user has keys added manually using ~
			home, err := os.UserHomeDir()
			if err == nil {
				key = strings.TrimPrefix(key, "~/")
				key = filepath.Join(home, key)
			}
		}
		if strings.HasPrefix(key, cfg.Ssh.KeyPath) {
			ownedKeys = append(ownedKeys, key)
		}
	}
	keyGenOptions := make([]string, 0)
	if len(cfg.Ssh.AcceptableKeyGenAlgorithms) > 0 {
		for _, keyAlg := range cfg.Ssh.AcceptableKeyGenAlgorithms {
			keyGenOptions = append(keyGenOptions, strings.ToUpper(keyAlg))
		}
	} else {
		keyGenOptions = append(keyGenOptions, config.RSA, config.ECDSA, config.ED25519)
	}
	var password string
	form := huh.NewForm(
		// replace old key step
		huh.NewGroup(
			huh.NewSelect[string]().
				Key(ROTATE_KEY_STR_KEY).
				Options(huh.NewOptions(ownedKeys...)...).
				Title("key to rotate"),
		),
		// key gen step
		huh.NewGroup(
			huh.NewSelect[string]().
				Key(KEY_GEN_ALGO_STR_KEY).
				Options(huh.NewOptions(keyGenOptions...)...).
				Title("Key Generation Method").
				Validate(func(s string) error {
					if len(s) == 0 {
						return fmt.Errorf("Must select a generation method")
					}
					return nil
				}),
			huh.NewInput().
				Key(KEY_GEN_PASSWORD).
				Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&password),
			huh.NewInput().
				Key(KEY_GEN_PASSWORD_VALIDATE).
				Title("Retype Password").
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if len(password) == 0 && len(s) != 0 {
						return fmt.Errorf("Password field is empty")
					}
					if s != password {
						return fmt.Errorf("Passwords do not match")
					}
					return nil
				}),
			huh.NewConfirm().
				Title("Rotate Key").
				Affirmative("Yes").
				Negative("No"),
		),
	).
		WithShowErrors(true).
		WithWidth(45).
		WithTheme(getTheme()).
		WithShowHelp(true)
	form.CancelCmd = func() tea.Msg {
		return abortedRotatedKeyForm{}
	}
	form.SubmitCmd = func() tea.Msg {
		keyGenType := form.GetString(KEY_GEN_ALGO_STR_KEY)
		password := form.GetString(KEY_GEN_PASSWORD)
		hostString := host
		keyToRotate := form.GetString(ROTATE_KEY_STR_KEY)
		if keyToRotate == "None" {
			keyToRotate = ""
		}
		keyPair, err := sshUtils.GenKey(hostString, keyGenType, password, cfg)
		return keyRotateRequest{
			host:       hostString,
			newKeySet:  keyPair,
			oldKeyPath: keyToRotate,
			err:        err,
		}
	}
	return KeyRotateModel{
		width:   45,
		height:  20,
		form:    form,
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
	}
}

func (rkm KeyRotateModel) Init() tea.Cmd {
	return rkm.form.Init()
}

func (rkm KeyRotateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		rkm.width = msg.Width
		rkm.height = msg.Height
		form, cmd := rkm.form.Update(msg)
		if updForm, ok := form.(*huh.Form); ok {
			rkm.form = updForm
		}
		return rkm, cmd
	case spinner.TickMsg:
		spin, cmd := rkm.spinner.Update(msg)
		rkm.spinner = spin
		return rkm, cmd
	}

	if rkm.form.State != huh.StateNormal {
		return rkm, nil
	}
	cmds := make([]tea.Cmd, 0)
	form, cmd := rkm.form.Update(msg)
	if updForm, ok := form.(*huh.Form); ok {
		rkm.form = updForm
	}
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	switch rkm.form.State {
	case huh.StateAborted:
		slog.Info("User aborted the rotate key form", "time", time.Now())
	case huh.StateCompleted:
		slog.Info("User submitted the rotate key form", "time", time.Now())
		cmds = append(cmds, rkm.spinner.Tick)
	}
	return rkm, tea.Batch(cmds...)
}

func (rkm KeyRotateModel) View() string {
	switch rkm.form.State {
	case huh.StateCompleted:
		return fmt.Sprintf("%s Generating keys...", rkm.spinner.View())
	case huh.StateAborted:
		return ""
	default:
		return rkm.form.View()
	}
}
