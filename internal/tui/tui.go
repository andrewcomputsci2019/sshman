package tui

import (
	"andrew/sshman/internal/config"
	"andrew/sshman/internal/sqlite"
	"andrew/sshman/internal/sshParser"
	"andrew/sshman/internal/sshUtils"
	"log/slog"
	"os"
	"os/exec"
	"time"

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
	host   string
	oldKey string
}

type removeOldKeyResult struct {
	host   string
	oldKey string
	err    error
}

type keyModalState struct {
	visible bool
	pubKey  string
	err     error
}

type AppModel struct {
	width, height int // this constitutes the entire terminal size
	// app components
	footer        FooterModel
	header        HeaderModel
	hostsModel    HostsPanelModel
	wizard        WizardViewModel
	keyForm       KeyGenModel
	keyRotateForm KeyRotateModel
	focusState    int
	db            *sqlite.HostDao
	pendingWrite  bool          // used to detect if write is needed before calling ssh process (only useful when writeThrough is disabled)
	cfg           config.Config // used to check if writeThrough is enabled if so forces
	sshOpts       []string
	keyModal      keyModalState
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
		args := []string{"-f", configPath}
		args = append(args, options...)
		args = append(args, host.Host)
		c = exec.Command("ssh", args...)
	} else {
		args := []string{"-f", configPath}
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
		}
		a.header.numberOfHost--
		// todo update ssh config file if write through enable and set pending write to true otherwise
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
		if a.keyModal.visible && (msg.String() == "esc" || msg.String() == "enter") {
			a.keyModal.visible = false
			return a, nil
		}
		if msg.Type == tea.KeyCtrlC {
			if a.pendingWrite {
				// todo serialize host config file
			}
			return a, tea.Quit
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
		model, cmd := a.wizard.Update(msg)
		a.wizard = model.(WizardViewModel)
		return a, cmd
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
		cmd := tea.ExecProcess(sshUtils.CopyKey(msg.newKeySet.PubKey, msg.host, a.cfg, a.sshOpts...),
			func(err error) tea.Msg {
				if err != nil {
					slog.Warn("Copy program failed to upload new key", "error", err, "host", msg.host, "public key", msg.newKeySet.PubKey)
					return abortedRotatedKeyForm{}
				}
				return removeOldKeyRequest{
					host:   msg.host,
					oldKey: msg.oldKeyPath,
				}
			})
		return a, cmd
	case removeOldKeyRequest:
		proc, err := sshUtils.RemoveOldKeyFromRemoteServer(msg.oldKey, msg.host, a.cfg, a.sshOpts...)
		if err != nil {
			slog.Warn("Failed to create a remote key removal script", "error", err)
			a.focusState = mainViewMode
			return a, nil
		}
		cmd := tea.ExecProcess(proc, func(err error) tea.Msg {
			return removeOldKeyResult{
				host:   msg.host,
				err:    err,
				oldKey: msg.oldKey,
			}
		})
		return a, cmd
	case removeOldKeyResult:
		a.focusState = mainViewMode
		if msg.err != nil {
			slog.Warn("failed to remove old key from remote server", "host", msg.host, "key", msg.oldKey, "error", msg.err)
			return a, nil
		}
		err := a.db.DeRegisterIdentityKeyFromHost(msg.host, msg.oldKey)
		if err != nil {
			slog.Warn("failed to delete key registered to host from database", "error", err, "host", msg.host, "key", msg.oldKey)
			return a, nil
		}
		err = os.Remove(msg.oldKey)
		if err != nil {
			slog.Warn("failed to delete key from disk", "file", msg.oldKey, "error", err)
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
			Render(a.keyForm.View())
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

func overlayView(base, modal string) string {
	return overlay.Composite(modal, base, overlay.Center, overlay.Center, 0, 0)
}
