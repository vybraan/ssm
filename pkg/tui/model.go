package tui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	lg "github.com/charmbracelet/lipgloss/v2"
	"github.com/lfaoro/ssm/pkg/sshconf"
)

type Model struct {
	config     *sshconf.Config
	showConfig bool

	li list.Model
	vp viewport.Model

	//external command
	ExtCmd    string
	ExitOnCmd bool
	ExitHost  string

	debug bool
	log   Log

	errbuf bytes.Buffer
	isDark bool
}

type ReloadConfigMsg struct{}
type ExitOnConnMsg struct{}
type FilterTagMsg struct {
	Arg string
}
type ShowConfigMsg struct{}

func NewModel(config *sshconf.Config, debug bool) *Model {
	m := &Model{}
	m.debug = debug
	m.config = config
	m.li = listFrom(config)
	m.log = NewLog(WithDebug(debug))
	m.ExtCmd = "ssh"
	m.vp = viewport.New()
	m.vp.SetWidth(40)
	m.vp.SetHeight(20)
	// m.vp.SetContent("test")
	return m
}

func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.SetWindowTitle("SSM | Secure Shell Manager"),
		// tea.SetBackgroundColor(color.Black),
		tea.RequestKeyboardEnhancements(),
		tea.EnterAltScreen,
		tea.EnableBracketedPaste,
		tea.EnableReportFocus,
		// tea.EnableMouseAllMotion,
		// tea.EnableMouseCellMotion,
	}
	if m.debug {
		cmds = append(cmds, AddLog("debug: isdarkbg %v", m.isDark))
	}
	m.li.NewStatusMessage(fmt.Sprintf("[%s]", m.ExtCmd))
	// reload config on edit
	cmds = append(cmds, m.watchCmd())
	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.isDark = msg.IsDark()
	case tea.WindowSizeMsg:
		m.li.SetSize(msg.Width, msg.Height-3)
		if m.debug {
			m.li.SetSize(msg.Width, msg.Height-9)
		}
		m.vp.SetHeight(m.li.Height())
		m.vp.SetWidth(msg.Width / 2)
	case ExitOnConnMsg:
		m.ExitOnCmd = true
		return m, nil
	case FilterTagMsg:
		m.li.SetFilterText(msg.Arg)
		m.li.SetFilteringEnabled(true)
		return m, nil
	case ReloadConfigMsg:
		m.li = listFrom(m.config)
		return m, nil
	case ShowConfigMsg:
		m.showConfig = true
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyTab:
			if m.ExtCmd == "ssh" {
				m.ExtCmd = "mosh"
				m.li.NewStatusMessage(fmt.Sprintf("[%s]", m.ExtCmd))
			} else {
				m.ExtCmd = "ssh"
				m.li.NewStatusMessage(fmt.Sprintf("[%s]", m.ExtCmd))
			}
		case tea.KeyEnter:
			// connect
			if m.li.FilterState() == list.Filtering {
				break
			}
			host, ok := m.li.SelectedItem().(item)
			if !ok {
				return m, AddError(fmt.Errorf("unable to find selected item: open bug report"))
			}
			if m.ExitOnCmd {
				m.ExitHost = host.title
				return m, tea.Quit
			}
			sshPath, err := exec.LookPath(m.ExtCmd)
			if err != nil {
				return m, AddError(fmt.Errorf("can't find `%s` cmd in your path: %v", m.ExtCmd, err))
			}
			var cmd *exec.Cmd
			cmd = exec.Command(sshPath, host.title)
			if host.title == "segfault.net" {
				_sshPath, err := exec.LookPath("sshpass")
				if err != nil {
					_sshPath, err = exec.LookPath("ssh")
					if err != nil {
						return m, AddError(fmt.Errorf("can't find `%s` cmd in your path: %v", m.ExtCmd, err))
					}
					cmd = exec.Command(_sshPath, "root@segfault.net")
				} else {
					cmd = exec.Command(_sshPath, "-p", "segfault", "ssh", "root@segfault.net")
				}
			}
			cmd.Stderr = &m.errbuf
			execmd := tea.ExecProcess(cmd, func(err error) tea.Msg {
				return tea.Batch(
					AddError(
						fmt.Errorf("connection closed: %v, err: %v", host.title, err),
					),
					AddError(fmt.Errorf("%s", m.errbuf.String())),
				)
			})
			return m, tea.Batch(
				execmd,
				AddError(fmt.Errorf("%s", m.errbuf.String())),
			)
		case tea.KeyEscape:
		case 'q':
			if m.li.FilterState() != 1 {
				return m, tea.Quit
			}
		}
		switch msg.Mod {
		// we're only interested in ctrl+<key>
		case tea.ModCtrl:
			switch msg.Code {
			case 'e':
				confFile := m.config.GetPath()
				cmd := exec.Command(os.Getenv("EDITOR"), confFile)
				cmd.Dir = filepath.Dir(confFile)
				cmd.Stderr = &m.errbuf
				execCmd := tea.ExecProcess(cmd, func(err error) tea.Msg {
					logCmd := AddLog("%v", err)
					var errCmd tea.Cmd
					if err != nil {
						errCmd = AddError(err)
					}
					return tea.Batch(logCmd, errCmd)
				})
				return m, tea.Sequence(
					execCmd,
					func() tea.Msg {
						return ReloadConfigMsg{}
					},
				)
			case 'r':
				return m, AddError(fmt.Errorf("run command on host: not yet implemented"))
			case 's':
				return m, AddError(fmt.Errorf("sftp: not yet implemented"))
			case 'v':
				m.showConfig = !m.showConfig
				m.setConfig()
			default:
				return m, AddError(fmt.Errorf("that's an interesting key combo! %s", msg))
			}
		default:
			cmds = append(cmds, ClearError())
		}
	}
	if len(m.errbuf.Bytes()) > 0 {
		cmds = append(cmds,
			AddError(fmt.Errorf("%v", m.errbuf.String())),
		)
		m.errbuf.Reset()
	}

	m.li, cmd = m.li.Update(msg)
	cmds = append(cmds, cmd)

	if m.showConfig {
		m.setConfig()
	}
	m.vp, cmd = m.vp.Update(msg)
	cmds = append(cmds, cmd)

	m.log, cmd = m.log.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) setConfig() {
	m.vp.SetContent("")
	i := m.li.GlobalIndex()
	host := m.config.Hosts[i]
	var out string
	keyStyle := lg.NewStyle().
		Foreground(lg.Color("#4682b4"))
	for i, k := range host.Options.Keys() {
		k = strings.ToTitle(k)
		k = keyStyle.Render(k)
		out += fmt.Sprintf("%s %s\n", k, host.Options.Values()[i])
	}
	m.vp.SetContent(out)
}

func (m *Model) View() string {
	var out string
	// style := lg.NewStyle().
	// 	Margin(1, 0, 0, 1).
	// 	Render
	// out += style(m.li.View())
	// out += style(m.log.View())
	vertView := lg.JoinVertical(0, m.li.View(), m.log.View())
	if m.debug {
		border := lg.NewStyle().Border(lg.RoundedBorder(), true)
		m.vp.Style = border
	}
	if m.showConfig {
		out += lg.JoinHorizontal(0.2, vertView, m.vp.View())
	} else {
		out += vertView
	}
	return out
}

func (m *Model) watchCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.config.Watch()
		if err != nil {
			return tea.Batch(
				AddError(err),
				AddLog("%v", err),
			)
		}
		return nil
	}
}
