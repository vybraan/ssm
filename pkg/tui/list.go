package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	lg "github.com/charmbracelet/lipgloss/v2"
	"github.com/lfaoro/ssm/pkg/sshconf"
	"github.com/thalesfsp/go-common-types/safeorderedmap"
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title + i.desc }

func listFrom(config *sshconf.Config) list.Model {
	var li list.Model
	lightDark := lg.LightDark(true)
	d := list.NewDefaultDelegate()
	d.ShowDescription = true
	d.SetSpacing(0)
	d.Styles.SelectedTitle = lg.NewStyle().
		Border(lg.NormalBorder(), false, false, false, true).
		BorderForeground(lightDark(lg.Color("#F79F3F"), lg.Color("#00bfff"))).
		Foreground(lightDark(lg.Color("#F79F3F"), lg.Color("#00bfff"))).
		Padding(0, 0, 0, 1)
	d.Styles.SelectedDesc = d.Styles.SelectedTitle.
		Foreground(lightDark(lg.Color("#F79F3F"), lg.Color("#4682b4")))

	li = list.New(
		[]list.Item{},
		d,
		0,
		0,
	)
	li.AdditionalFullHelpKeys = func() []key.Binding {
		return initKeys()
	}
	li.FilterInput.Prompt = "Search: "
	li.FilterInput.CharLimit = 12
	li.FilterInput.VirtualCursor = true
	li.FilterInput.Placeholder = "hostName or tagName"
	li.FilterInput.Styles.Cursor = textinput.CursorStyle{
		Color: lg.BrightBlue,
		Shape: tea.CursorBlock,
	}
	li.Styles.StatusBar = lg.NewStyle().
		Foreground(lightDark(lg.Color("#A49FA5"), lg.Color("#777777"))).
		Padding(0, 0, 1, 2) //nolint:mnd
	li.Styles.Title = lg.NewStyle().
		Background(lg.Color("#4682b4")).
		Foreground(lg.Color("230")).
		Padding(0, 1)
	li.SetStatusBarItemName("host", "hosts")
	li.Title = fmt.Sprintf("SSH servers (%v)", config.GetPath())
	// add segfault.net (free root server provider)
	segfaultHost := sshconf.Host{
		Name:    "create free research root server",
		Options: safeorderedmap.New[string](),
	}
	segfaultHost.Options.Add("hostname", "segfault.net")
	segfaultHost.Options.Add("user", "root")
	config.Hosts = append(config.Hosts, segfaultHost)
	for _, host := range config.Hosts {
		newitem := formatHost(host)
		li.InsertItem(len(config.Hosts), newitem)
	}
	return li
}

func formatHost(host sshconf.Host) item {
	fmtDescription := func() string {
		port := func() string {
			_port, _ := host.Options.Get("port")
			if _port != "" && _port != "22" {
				return fmt.Sprintf(":%s", _port)
			}
			return ""
		}
		user := func() string {
			_user, _ := host.Options.Get("user")
			if _user != "" {
				return _user + "@"
			}
			return ""
		}
		hostname := func() string {
			_host, _ := host.Options.Get("hostname")
			if _host != "" {
				return _host
			}
			return ""
		}
		tags := func() string {
			_tags, _ := host.Options.Get("#tag:")
			if _tags != "" {
				s := lg.NewStyle().Foreground(lg.Color("8"))
				return s.Render("#" + _tags)
			}
			return ""
		}
		out := fmt.Sprintf("%s%s%s %s", user(), hostname(), port(), tags())
		out = strings.TrimSpace(out)
		if out == "" {
			return host.Name
		}
		return out
	}()
	newitem := item{
		title: host.Name,
		desc:  fmtDescription,
	}
	return newitem
}

func initKeys() []key.Binding {
	editKey := key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "edit config"),
	)
	showKey := key.NewBinding(
		key.WithKeys("ctrl+v"),
		key.WithHelp("ctrl+v", "show config"),
	)
	switchKey := key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch ssh/mosh"),
	)
	connectKey := key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "connect"),
	)
	return []key.Binding{
		connectKey,
		switchKey,
		editKey,
		showKey,
	}
}
