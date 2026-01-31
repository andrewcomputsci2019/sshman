package tui

import (
	"andrew/sshman/internal/config"
	"andrew/sshman/internal/sqlite"
	"andrew/sshman/internal/sshParser"
	"andrew/sshman/internal/sshUtils"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

const (
	mainViewMode int = iota // user on main screen
	wizardMode              // user is in entry wizard mode adding a host
	keyGenForm
	rotateKeyGenForm
)

const (
	totalBorderHeight = 3
)

type updateHostsMessage struct {
	host sqlite.Host
}

type newHostsMessage struct {
	host sqlite.Host
}

type userAddHostMessage struct{}

// user doesn't want to add a new host
type userExitWizard struct{}

type deleteHostMessage struct {
	host string
}

type sshProcFinished struct {
	err error
}

type pingResult struct {
	host          string
	hostReachable bool
	ping          time.Duration
	err           error
}

type startKeyRotateForm struct {
	host string
}

type startKeyGenerationForm struct {
	host string
}

type removeOldKeyRequest struct {
	host       string
	oldKey     string
	newKeyPair sshUtils.KeyPair
	err        error
}

type removeOldKeyResult struct {
	host          string
	oldKey        string
	newKeyPair    sshUtils.KeyPair
	err           error
	keyWasRemoved bool // shows wether the script ran or not
	statusMsg     string
}

type failedToCopyKey struct {
	pair sshUtils.KeyPair
	err  error
}

type keyModalState struct {
	visible bool // focus state
	pubKey  string
	err     error
}

// todo hook up modal to show when rotating keys to get user consent to script usage

type rotateKeyRemoveModalState struct {
	visible     bool // focus state
	err         error
	host        string
	keyPair     sshUtils.KeyPair
	keyToRemove string
	scriptView  viewport.Model
	script      string
	cmd         tea.Cmd
}

type rotateKeyCopyModalState struct {
	visible bool // focus state
	err     error
	host    string
	keys    sshUtils.KeyPair
	oldKey  string
	cmd     tea.Cmd
}

type rotateKeyResultModal struct {
	visible bool // focus state
	err     error
	message string
}

type failedToCopyModal struct {
	visible bool
	pair    sshUtils.KeyPair
	err     error
}

type AppModel struct {
	width, height int // this constitutes the entire terminal size
	// app components
	footer                FooterModel
	header                HeaderModel
	hostsModel            HostsPanelModel
	wizard                WizardViewModel
	keyForm               KeyGenModel
	keyRotateForm         KeyRotateModel
	focusState            int
	db                    *sqlite.HostDao
	pendingWrite          bool          // used to detect if write is needed before calling ssh process (only useful when writeThrough is disabled)
	cfg                   config.Config // used to check if writeThrough is enabled if so forces
	sshOpts               []string
	keyModal              keyModalState
	rotateCopyModal       rotateKeyCopyModalState
	rotateRemoveKeyModal  rotateKeyRemoveModalState
	rotateResultModal     rotateKeyResultModal
	rotateCopyFailedModal failedToCopyModal
}

// todo implement model func

func (a AppModel) Init() tea.Cmd {
	return nil
}

func getWriteThroughOption(configuration *bool) bool {
	// return true if option not set or explicitly turned on and false otherwise
	return configuration == nil || *configuration
}

func runSSHProgram(host sqlite.Host, sshPath string, configPath string, options ...string) tea.Cmd {
	// note options need to be passed with a prefix of -o
	var c *exec.Cmd
	if sshPath == "" {
		args := []string{"-F", configPath}
		args = append(args, options...)
		args = append(args, host.Host)
		c = exec.Command("ssh", args...)
	} else {
		args := []string{"-F", configPath}
		args = append(args, options...)
		args = append(args, host.Host)
		c = exec.Command(sshPath, args...)
	}
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return sshProcFinished{err: err}
	})
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		width := a.width
		a.header.width = msg.Width
		foot, _ := a.footer.Update(msg)
		a.footer = foot.(FooterModel)
		remainingHeight := a.height - a.header.height - a.footer.height - totalBorderHeight // (-2 as header as bottom border and footer has a top border)
		newDim := tea.WindowSizeMsg{Width: width, Height: remainingHeight}                  // (-2 on width as we want side borders)
		hostModel, cmd := a.hostsModel.Update(newDim)
		a.hostsModel = hostModel.(HostsPanelModel)
		wiz, _ := a.wizard.Update(newDim)
		a.wizard = wiz.(WizardViewModel)
		a.keyForm.width = msg.Width
		a.keyForm.height = newDim.Height
		return a, cmd
	case userAddHostMessage:
		// Show wizard state, and create a new wizard with current dimensions of viewport
		a.focusState = wizardMode
		newWiz, _ := NewWizardViewModel().Update(tea.WindowSizeMsg{Height: a.wizard.height, Width: a.wizard.width})
		a.wizard = newWiz.(WizardViewModel)
		return a, nil
	case userExitWizard: // leave sshWizard view
		a.focusState = mainViewMode
		return a, nil
	case deleteHostMessage:
		err := a.db.Delete(sqlite.Host{Host: msg.host})
		if err != nil {
			slog.Error("Failed to delete host from db", "Host", msg.host)
			return a, nil
		}
		a.header.numberOfHost--
		// todo update ssh config file if write through enable and set pending write to true otherwise
		hosts, err := a.db.GetAll()
		if err != nil {
			slog.Error("Failed to get all host after db modification")
			return a, nil
		}
		if getWriteThroughOption(a.cfg.StorageConf.WriteThrough) {
			sshParser.SerializeHostToFile(a.cfg.GetSshConfigFilePath(), hosts)
		} else {
			a.pendingWrite = true
		}
		a.hostsModel.data = hosts
		a.hostsModel.refreshTableRows()
		return a, nil
	case newHostsMessage:
		// todo insert new host and then get updated table
		a.focusState = mainViewMode
		newHost := msg.host
		err := a.db.Insert(newHost)
		if err != nil {
			slog.Error("Failed to insert new host into the db", "error", err)
			return a, nil
		}
		a.header.numberOfHost++
		if getWriteThroughOption(a.cfg.StorageConf.WriteThrough) {
			err = sshParser.AddHostToFile(a.cfg.GetSshConfigFilePath(), newHost)
		} else {
			a.pendingWrite = true // mark that a full dump into ssh file will be required before connecting
		}
		if err != nil {
			slog.Error("Failed to write host into ssh config file", "host", newHost, "path", a.cfg.GetSshConfigFilePath())
		}
		model, cmd := a.hostsModel.Update(msg)
		a.hostsModel = model.(HostsPanelModel)
		return a, cmd

	case updateHostsMessage:
		// todo upsert a host into database, no need to do anything else after, tui will update it self on next msg anyways
		err := a.db.Update(msg.host)
		if err != nil {
			slog.Error("Failed to update host into db", "host", msg, "error", err)
			return a, nil
		}
		if getWriteThroughOption(a.cfg.StorageConf.WriteThrough) {
			hosts, err := a.db.GetAll()
			if err != nil {
				a.pendingWrite = true
				slog.Error("failed to fetch ssh hosts")
				return a, nil
			}
			err = sshParser.SerializeHostToFile(a.cfg.GetSshConfigFilePath(), hosts)
			if err != nil {
				slog.Error("Failed to serialize the host into the ssh config file", "error", err)
				a.pendingWrite = true
			}
		} else {
			a.pendingWrite = true
		}
		return a, nil
	case connectHostMessage:
		timeStamp := time.Now()
		err := a.db.UpdateLastConnection(msg.host.Host, &timeStamp)
		if err != nil {
			slog.Warn("Failed to update last connection timestamp", "Hosts", msg.host, "Error", err)
		} else {
			a.hostsModel.updateLastConnection(msg.host.Host, timeStamp)
		}
		if a.pendingWrite {
			// todo write out host from db into serialization file
			// to do this get all Host from database and then serialize them all back into the config file
			hosts, err := a.db.GetAll()
			if err != nil {
				slog.Error("Failed to get hosts from database", "error", err)
			} else {
				err = sshParser.SerializeHostToFile(a.cfg.GetSshConfigFilePath(), hosts)
				if err != nil {
					slog.Error("Failed to serialize host into ssh config file", "file", a.cfg.GetSshConfigFilePath(), "error", err)
				}
			}
			a.pendingWrite = !a.pendingWrite
		}
		// todo in the future we also want to pass options given from the command line
		exeCommand := runSSHProgram(msg.host, a.cfg.Ssh.ExcPath, a.cfg.GetSshConfigFilePath())
		return a, exeCommand

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			if a.pendingWrite {
				// todo serialize host config file
			}
			return a, tea.Quit
		}
		// check if any modal is visible
		// going to be honest i think resetting the focus state is not needed since its adjusted in the modal
		// creation but im going to leave it in case something changes in future and that isn't true
		if a.keyModal.visible {
			if msg.String() == "esc" || msg.String() == "enter" {
				a.keyModal.visible = false
				a.focusState = mainViewMode
			}
			return a, nil
		} else if a.rotateCopyModal.visible { // todo implement these blocks that handle modal views
			if msg.String() == "esc" {
				a.rotateCopyModal.visible = false
				a.focusState = mainViewMode
				if a.rotateCopyModal.err == nil {
					// show results screen with pub and private key loc
					status := "User canceled copy request\n"
					newKeys := "New keys can be found: " + a.rotateCopyModal.keys.PrivateKey + "(.pub)\n"
					oldKeys := "To remove the old key from remote server look for the comment that matches the private key"
					a.rotateResultModal = rotateKeyResultModal{
						err:     nil,
						message: fmt.Sprintf("%s%s%s", status, newKeys, oldKeys),
						visible: true,
					}
				}
			} else if msg.String() == "enter" {
				if a.rotateCopyModal.err == nil {
					a.rotateCopyModal.visible = false
					a.focusState = mainViewMode
					return a, a.rotateCopyModal.cmd
				} else { // if error is not nil accept enter as esc, this was because an error was generated producing the keys
					a.rotateCopyModal.visible = false
					a.focusState = mainViewMode
				}
			}
			return a, nil
		} else if a.rotateCopyFailedModal.visible {
			if msg.String() == "esc" || msg.String() == "enter" {
				a.rotateCopyFailedModal.visible = false
				a.focusState = mainViewMode
			}

		} else if a.rotateRemoveKeyModal.visible { // this will need extra work for handling viewport navigation
			if a.rotateRemoveKeyModal.err != nil && (msg.String() == "enter" || msg.String() == "esc") {
				a.rotateRemoveKeyModal.visible = false
				a.focusState = mainViewMode
				return a, nil
			}
			if msg.String() == "enter" {
				a.rotateRemoveKeyModal.visible = false
				a.focusState = mainViewMode
				return a, a.rotateRemoveKeyModal.cmd
			}
			if msg.String() == "esc" {
				// here we should show the results modal
				// should say that the key has been copied but the old key has not been removed
				a.rotateRemoveKeyModal.visible = false
				statusMsg := "New Key %s was copied to remote.\nOld key %s was not removed from remote and should be removed"
				statusMsg = fmt.Sprintf(statusMsg, filepath.Base(a.rotateRemoveKeyModal.keyPair.PubKey), filepath.Base(a.rotateRemoveKeyModal.keyToRemove))
				a.rotateResultModal = rotateKeyResultModal{
					visible: true,
					err:     nil,
					message: statusMsg,
				}
			}
			// view port handling
			if msg.String() == "left" || msg.String() == "h" || msg.String() == "right" || msg.String() == "l" {
				m, cmd := a.rotateRemoveKeyModal.scriptView.Update(msg)
				a.rotateRemoveKeyModal.scriptView = m
				return a, cmd
			}
			if msg.String() == "k" || msg.String() == "up" {
				if a.rotateRemoveKeyModal.scriptView.YOffset >= 1 {
					a.rotateRemoveKeyModal.scriptView.YOffset--
				}
				return a, nil
			}
			if msg.String() == "i" || msg.String() == "down" {
				a.rotateRemoveKeyModal.scriptView.YOffset++
				return a, nil
			}
			return a, nil

		} else if a.rotateResultModal.visible { // enter and esc being only options here
			if msg.String() == "esc" || msg.String() == "enter" {
				a.rotateResultModal.visible = false
				a.focusState = mainViewMode
			}
			return a, nil
		}
		if a.focusState == mainViewMode {
			model, cmd := a.hostsModel.Update(msg)
			a.hostsModel = model.(HostsPanelModel)
			return a, cmd
		}
		if a.focusState == keyGenForm {
			model, cmd := a.keyForm.Update(msg)
			a.keyForm = model.(KeyGenModel)
			return a, cmd
		}
		if a.focusState == rotateKeyGenForm {
			model, cmd := a.keyRotateForm.Update(msg)
			a.keyRotateForm = model.(KeyRotateModel)
			return a, cmd
		}
		if a.focusState == wizardMode {
			model, cmd := a.wizard.Update(msg)
			a.wizard = model.(WizardViewModel)
			return a, cmd
		}
		return a, nil
	case sshProcFinished:
		if msg.err != nil {
			slog.Error("ssh ran into an error", "error", msg.err)
		}
		return a, nil
	case pingResult:
		update, cmd := a.hostsModel.Update(msg)
		a.hostsModel = update.(HostsPanelModel)
		return a, cmd
	case abortedKeyGenForm:
		a.focusState = mainViewMode
		return a, nil
	case abortedRotatedKeyForm:
		a.focusState = mainViewMode
		return a, nil
	// todo set focus state and initialize forms according to the host given
	case startKeyGenerationForm:
		a.focusState = keyGenForm
		a.keyForm = NewKeyGenModel(msg.host, a.cfg)
		return a, a.keyForm.Init()
	case startKeyRotateForm:
		a.focusState = rotateKeyGenForm
		hostsKeys, err := a.db.GetAllHostsIdentityKeys(msg.host)
		if err != nil {
			slog.Warn("Failed to get hosts keys due to error", "Host", msg.host, "error", err)
			return a, nil
		}
		a.keyRotateForm = NewKeyRotateModel(msg.host, hostsKeys, a.cfg)
		return a, a.keyRotateForm.Init()
	case keyGenResult:
		a.focusState = mainViewMode
		a.keyModal = keyModalState{
			visible: true,
			pubKey:  msg.keyPair.PubKey,
			err:     msg.err,
		}
		if msg.err == nil {
			err := a.db.RegisterNewIdentityKeyForHost(msg.host, msg.keyPair.PrivateKey)
			if err == nil {
				if getWriteThroughOption(a.cfg.StorageConf.WriteThrough) {
					hosts, err := a.db.GetAll()
					if err != nil {
						a.pendingWrite = true
						slog.Warn("Failed to get all hosts from database", "error", err)
						return a, nil
					}
					a.hostsModel.data = hosts
					a.hostsModel.refreshTableRows()
					err = sshParser.SerializeHostToFile(a.cfg.GetSshConfigFilePath(), hosts)
					if err != nil {
						slog.Warn("Failed to serialize hosts into ssh config file", "error", err)
					}
				} else {
					a.pendingWrite = true
					updHost, err := a.db.Get(msg.host)
					if err == nil {
						updHosts, cmd := a.hostsModel.Update(updateHostsMessage{host: updHost})
						a.hostsModel = updHosts.(HostsPanelModel)
						return a, cmd
					} else {
						slog.Warn("failed to get host after issuing a new identity key", "host", msg.host, "err", err)
					}
				}
			} else {
				slog.Error("Failed to save generated key to sqlite database", "host", msg.host, "error", err)
			}
		} else {
			slog.Warn("received an error with key gen result", "error", msg.err)
		}
		return a, nil
	case keyRotateRequest:
		// todo show modal then okay the request if user accepts
		// we should still create the exec object now, to show its string representation
		// in the modal
		// the rest of this needs to be hookup to another message like keyRotateMsg
		// enter to accept
		// esc to cancel
		// if the user exit here don't remove the old key from the host

		a.rotateCopyModal = newRotateCopyModal(msg)
		if msg.err == nil {
			err := a.db.RegisterNewIdentityKeyForHost(msg.host, msg.newKeySet.PrivateKey)
			if err != nil {
				a.rotateCopyModal.err = err
				return a, nil
			}
			cmd := tea.ExecProcess(sshUtils.CopyKey(msg.newKeySet.PubKey, msg.host, a.cfg, a.sshOpts...),
				func(err error) tea.Msg {
					if err != nil {
						slog.Warn("Copy program failed to upload new key", "error", err, "host", msg.host, "public key", msg.newKeySet.PubKey)
						return failedToCopyKey{
							err:  err,
							pair: msg.newKeySet,
						}
					}
					if a.cfg.Ssh.RemovePubKeyAfterGen {
						err := os.Remove(msg.newKeySet.PubKey)
						if err != nil {
							slog.Warn(
								"Failed to remove pub key from keystore after copy", "PubKey", msg.newKeySet.PubKey,
								"err", err,
							)
						}
					}
					return removeOldKeyRequest{
						host:       msg.host,
						oldKey:     msg.oldKeyPath,
						newKeyPair: msg.newKeySet,
					}
				})
			updHost, err := a.db.Get(msg.host)
			if err != nil {
				slog.Warn("Failed to fetch updated host after key addition in key rotation")
			} else {
				m, _ := a.hostsModel.Update(updateHostsMessage{
					host: updHost,
				})
				a.hostsModel = m.(HostsPanelModel)
			}
			a.rotateCopyModal.cmd = cmd
			return a, nil
		}
		return a, nil
	case failedToCopyKey:
		a.rotateCopyFailedModal = newFailedToCopyModal(msg)
		return a, nil
	case removeOldKeyRequest:
		// ask user if they want to continue with the request
		// the modal should show the key name being removed
		// the modal should also show the script using a viewport so the user can scroll through what the script is going to run
		// enter accept
		// esc cancel --> in this case send a removeOldKeyResult msg so the old key is at least de registered from host file

		if msg.oldKey == "" { // case where users does not want to remove old key
			a.rotateResultModal = rotateKeyResultModal{
				visible: true,
				err:     nil,
				message: "Key uploaded to remote, no key to remove operation complete",
			}
			return a, nil
		}
		proc, err := sshUtils.RemoveOldKeyFromRemoteServer(msg.oldKey, msg.host, a.cfg, a.sshOpts...)
		if err != nil {
			slog.Warn("Failed to create a remote key removal script", "error", err)
			a.focusState = mainViewMode
			a.rotateResultModal = newRotateResultModal(removeOldKeyResult{
				err:           err,
				oldKey:        msg.oldKey,
				newKeyPair:    msg.newKeyPair,
				keyWasRemoved: false,
				statusMsg:     "Failed to generate removal script, likely due to key not being managed by ssh_man",
			})
			return a, nil
		}
		cmd := tea.ExecProcess(proc, func(err error) tea.Msg {
			return removeOldKeyResult{
				host:          msg.host,
				err:           err,
				oldKey:        msg.oldKey,
				newKeyPair:    msg.newKeyPair,
				keyWasRemoved: err == nil,
			}
		})
		a.rotateRemoveKeyModal = newRotateRemoveModal(msg)
		a.rotateRemoveKeyModal.cmd = cmd
		a.rotateRemoveKeyModal.scriptView.SetHorizontalStep(5)
		a.rotateRemoveKeyModal.scriptView.Height = 6
		a.rotateRemoveKeyModal.script = proc.String()
		return a, nil
	case removeOldKeyResult:
		a.focusState = mainViewMode
		a.rotateResultModal = newRotateResultModal(msg)
		if msg.err != nil {
			slog.Warn("failed to remove old key from remote server", "host", msg.host, "key", msg.oldKey, "error", msg.err)
			a.rotateResultModal.message = "Failed to remove old key from server due to error: " + msg.err.Error()
			return a, nil
		}
		err := a.db.DeRegisterIdentityKeyFromHost(msg.host, msg.oldKey)
		if err != nil {
			slog.Warn("failed to delete key registered to host from database", "error", err, "host", msg.host, "key", msg.oldKey)
			a.rotateResultModal.message = "New key was uploaded to server but failed to remove old one from database, error: " + err.Error()
			return a, nil
		}
		host, err := a.db.Get(msg.host)
		if err != nil {
			slog.Warn("tui will be out of date with sqlite backend")
		} else {
			m, _ := a.hostsModel.Update(updateHostsMessage{
				host: host,
			})
			a.hostsModel = m.(HostsPanelModel)
		}
		if msg.keyWasRemoved {
			err = os.Remove(msg.oldKey)
			if err != nil {
				slog.Warn("failed to delete key from disk", "file", msg.oldKey, "error", err)
				a.rotateResultModal.message = "New key was uploaded but failed to remove old key: " + filepath.Base(msg.oldKey) + "\n from disk"
			} else {
				a.rotateResultModal.message = "New Key was uploaded and old key was removed from config and remote server"
			}
		} else {
			a.rotateResultModal.message = fmt.Sprintf("New Key %s was uploaded\nOld Key %s was not removed from remote", filepath.Base(msg.newKeyPair.PubKey), filepath.Base(msg.oldKey))
		}
		return a, nil
	default:
		switch a.focusState {
		case keyGenForm:
			updForm, cmd := a.keyForm.Update(msg)
			a.keyForm = updForm.(KeyGenModel)
			return a, cmd
		case rotateKeyGenForm:
			updForm, cmd := a.keyRotateForm.Update(msg)
			a.keyRotateForm = updForm.(KeyRotateModel)
			return a, cmd
		default:
			return a, nil
		}
	}
}

func (a AppModel) View() string {
	a.footer.currentKeymap = a.hostsModel
	// todo render header with a bottom normal border
	header := lipgloss.NewStyle().Width(a.width).
		Height(a.header.Height()).
		Border(lipgloss.NormalBorder(), false, false, true).
		Align(lipgloss.Left).
		Render(a.header.View())
	// todo HostModel should have side borders
	var center string
	switch a.focusState {
	case mainViewMode:
		center = lipgloss.NewStyle().Width(a.width).
			Height(a.hostsModel.height).
			Render(a.hostsModel.View())
	case keyGenForm:
		center = lipgloss.NewStyle().Width(a.width).
			Height(a.keyForm.height).
			Align(lipgloss.Center).
			Render(a.keyForm.View())
	case rotateKeyGenForm:
		center = lipgloss.NewStyle().Width(a.width).
			Height(a.keyForm.height).
			Align(lipgloss.Center).
			Render(a.keyRotateForm.View())
	default:
		center = lipgloss.NewStyle().Width(a.width).
			Height(a.wizard.height).
			Align(lipgloss.Center).
			Render(a.wizard.View())
	}
	footer := lipgloss.NewStyle().
		Width(a.width).
		Height(a.footer.Height()).
		Border(lipgloss.NormalBorder(), true, false, false).
		Align(lipgloss.Left).
		Render(a.footer.View())
	base := lipgloss.JoinVertical(lipgloss.Left, header, center, footer)
	if a.keyModal.visible {
		dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(base)
		return overlayView(dimmed, a.keyModalView())
	}
	if a.rotateCopyModal.visible {
		dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(base)
		return overlayView(dimmed, a.rotateKeyCopyModalView())
	}
	if a.rotateCopyFailedModal.visible {
		dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(base)
		return overlayView(dimmed, a.rotateFailedToCopyKeyModalView())
	}
	if a.rotateRemoveKeyModal.visible {
		dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(base)
		return overlayView(dimmed, a.rotateKeyRemoveModalView())
	}
	if a.rotateResultModal.visible {
		dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(base)
		return overlayView(dimmed, a.rotateKeyResultView())
	}
	return base
}

func NewAppModel(hosts []sqlite.Host, db *sqlite.HostDao, cfg config.Config, sshOpts ...string) AppModel {
	options := make([]string, 0)
	for _, opt := range sshOpts {
		options = append(options, "-o "+opt)
	}
	appModel := AppModel{
		db:         db,
		header:     NewHeaderModel(uint(len(hosts))),
		footer:     NewFooterModel(),
		focusState: int(mainViewMode),
		hostsModel: NewHostsPanelModel(cfg, hosts),
		wizard:     NewWizardViewModel(),
		sshOpts:    options,
		cfg:        cfg,
	}
	appModel.footer.currentKeymap = appModel.hostsModel
	appModel.rotateRemoveKeyModal.scriptView = viewport.New(60, 15)
	return appModel
}

func (a AppModel) keyModalView() string {
	title := lipgloss.NewStyle().Bold(true).Render("Key Generated")
	body := "Public key saved at:\n" + a.keyModal.pubKey
	if a.keyModal.err != nil {
		body = "Key generation failed:\n" + a.keyModal.err.Error()
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(2, 2).
		Width(max(a.width/2, 60)).
		Render(title + "\n\n" + body + "\n\nPress enter or esc to close")
}

func (a AppModel) rotateKeyCopyModalView() string {
	width := max(60, a.width/2)
	title := lipgloss.NewStyle().Bold(true).Render("Copy Key To Remote")
	msg := fmt.Sprintf("Do you wish to copy key %s to remote %s", filepath.Base(a.rotateCopyModal.keys.PubKey), a.rotateCopyModal.host)
	tail := lipgloss.NewStyle().Bold(true).Render("Press enter to proceeded or Esc to cancel")
	if a.rotateCopyModal.err != nil {
		msg = fmt.Sprintf("Cannot Copy key to remote for reason %s", a.rotateCopyModal.err)
		tail = "Press enter or esc to close"
	}

	content := fmt.Sprintf("%s\n%s\n\n%s\n", title, msg, tail)
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).
		Width(width).
		Padding(2, 2).
		Render(content)
}

func (a AppModel) rotateKeyRemoveModalView() string {
	width := max(70, a.width/2)
	a.rotateRemoveKeyModal.scriptView.Width = width
	title := lipgloss.NewStyle().Bold(true).Render("Remove Key From Remote")
	var content string
	var tail string
	if a.rotateRemoveKeyModal.err != nil { //
		content = "Error encountered. Error: " + a.rotateRemoveKeyModal.err.Error()
		tail = "\nPress enter or esc to close"
	} else {
		// content = a.rotateRemoveKeyModal.script
		content = lipgloss.NewStyle().Width(width - 2).Render(a.rotateRemoveKeyModal.script)
		tail = "\nPress enter to continue or esc to cancel"
	}
	a.rotateRemoveKeyModal.scriptView.SetContent(content)
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(width).
		Padding(2, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, a.rotateRemoveKeyModal.scriptView.View()), tail)
}

func (a AppModel) rotateKeyResultView() string {
	width := max(60, a.width/2)
	title := lipgloss.NewStyle().Bold(true).Render("Key Rotation Result")
	topSeparator := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("/", width-4))
	bottomSeparator := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("\\", width-4))
	content := fmt.Sprintf("%s\n%s\n%s\n%s\n\nPress enter or esc to close", title, topSeparator, a.rotateResultModal.message, bottomSeparator)

	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).
		Width(width).
		Padding(2, 2).
		Render(content)
}

func (a AppModel) rotateFailedToCopyKeyModalView() string {
	width := max(60, a.width/2)
	title := lipgloss.NewStyle().Bold(true).Render("Copy Failed")
	errMsg := "Encountered error: " + a.rotateCopyFailedModal.err.Error()
	content := "Key(s) Location: " + a.rotateCopyFailedModal.pair.PrivateKey + "(.pub)"
	tail := "Press enter or esc to close"
	base := fmt.Sprintf("%s\n%s\n%s\n%s", title, errMsg, content, tail)
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(width).
		Padding(2, 2).
		Render(base)
}

func overlayView(base, modal string) string {
	return overlay.Composite(modal, base, overlay.Center, overlay.Center, 0, 0)
}

func overlayViewWithOffsets(base, modal string, xOff, yOff int) string {
	return overlay.Composite(modal, base, overlay.Center, overlay.Center, xOff, yOff)
}

// modal related code
func newRotateCopyModal(req keyRotateRequest) rotateKeyCopyModalState {
	return rotateKeyCopyModalState{
		visible: true,
		err:     req.err,
		host:    req.host,
		oldKey:  req.oldKeyPath,
		keys:    req.newKeySet,
	}
}

func newRotateRemoveModal(req removeOldKeyRequest) rotateKeyRemoveModalState {
	return rotateKeyRemoveModalState{
		visible:     true,
		err:         req.err,
		host:        req.host,
		keyPair:     req.newKeyPair,
		keyToRemove: req.oldKey,
		scriptView:  viewport.New(30, 6),
	}
}

func newRotateResultModal(req removeOldKeyResult) rotateKeyResultModal {
	return rotateKeyResultModal{
		visible: true,
		err:     req.err,
		message: req.statusMsg,
	}
}

func newFailedToCopyModal(req failedToCopyKey) failedToCopyModal {
	return failedToCopyModal{
		err:     req.err,
		visible: true,
		pair:    req.pair,
	}
}

// helpers for debugging ui layouts for modals

// showKeyCopyModal debug util for checking if the modal ui is displayed correctly
//
// err should be nil if the modal should be displayed normally and
// non nil if the modal should show an error instead
func (a *AppModel) ShowKeyCopyModal(err error) {
	// todo show modal with faux data
	if !a.cfg.DevMode {
		return
	}
	a.rotateCopyModal = rotateKeyCopyModalState{
		visible: true,
		err:     err,
		host:    "DEBUG LAYOUT",
		keys: sshUtils.KeyPair{
			PrivateKey: filepath.Join(a.cfg.Ssh.KeyPath, "debug_test"),
			PubKey:     filepath.Join(a.cfg.Ssh.KeyPath, "debug_test.pub"),
		},
		oldKey: "Just A test",
		cmd:    tea.SetWindowTitle("Key Copy inited next step"),
	}
}

// showKeyFailedCopyModal debug util for checking if the modal ui is displayed correctly
//
// err should be supplied and non nil
func (a *AppModel) ShowKeyFailedCopyModal(err error) {
	// todo show modal with
	if !a.cfg.DevMode || err == nil {
		return
	}

	a.rotateCopyFailedModal = failedToCopyModal{
		visible: true,
		pair: sshUtils.KeyPair{
			PrivateKey: filepath.Join(a.cfg.Ssh.KeyPath, "debug_test"),
			PubKey:     filepath.Join(a.cfg.Ssh.KeyPath, "debug_test.pub"),
		},
		err: err,
	}
}

// showKeyRemoveModal debug util for checking if the modal ui is displayed correctly
//
// err should be non nil if the modal should display an error
func (a *AppModel) ShowKeyRemoveModal(err error) {
	if !a.cfg.DevMode {
		return
	}
	sampleScriptText := "This is some long text that will test the viewport ability to render content\nthat spans many columns this text will have many rows as well to test scalability,\nif the content does not wrap correctly then something is wrong The viewport uses default key binds provided by cham these include left,right,h,l,up,down,i,k. \nIdeally this content should display cleanly and clearly allowing a user to see a script before it runs; and make the informed choice of wether they want to continue to run the script after inspection."

	a.rotateRemoveKeyModal = rotateKeyRemoveModalState{
		visible: true,
		err:     err,
		host:    "myHost",
		keyPair: sshUtils.KeyPair{
			PrivateKey: filepath.Join(a.cfg.Ssh.KeyPath, "debug_test"),
			PubKey:     filepath.Join(a.cfg.Ssh.KeyPath, "debug_test.pub"),
		},
		keyToRemove: "TEST_KEY",
		scriptView:  viewport.New(60, 6),
		script:      sampleScriptText,
		cmd:         tea.SetWindowTitle("rotateKeyRemoveModel inited next step"),
	}
	a.rotateRemoveKeyModal.scriptView.SetHorizontalStep(5)
}

// showKeyRemoveResultModal debug util for checking if the modal ui is displayed correctly
//
// err should be non nil if the modal should display an error
func (a *AppModel) ShowKeyRemoveResultModal(err error) {
	if !a.cfg.DevMode {
		return
	}
	a.rotateResultModal = rotateKeyResultModal{
		visible: true,
		err:     err,
		message: `This is a test message, where this would be a status message on the outcome of the rotate key automation
		event. An example would be User canceled the remove key event, the new key was copied but the old key was not removed from
		the remote`,
	}
}

func (a *AppModel) ShowKeyGenModal(err error) {
	if !a.cfg.DevMode {
		return
	}
	if err != nil {
		a.keyModal = keyModalState{
			visible: true,
			err:     err,
		}
	} else {
		a.keyModal = keyModalState{
			visible: true,
			pubKey:  "example pub key would go on this line",
		}
	}
}
