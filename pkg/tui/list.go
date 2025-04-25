package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
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

	li = list.New(
		[]list.Item{},
		d,
		0,
		0,
	)
	li.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			connectKey,
			switchKey,
			editKey,
			showKey,
		}
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
	for _, host := range config.Hosts {
		if host.Name == "*" {
			continue
		}
		fmtDescription := func() string {
			port := func() string {
				_port, _ := host.Options.Get("port")
				if _port != "" {
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
			return out
		}()
		newitem := item{
			title: host.Name,
			desc:  fmtDescription,
		}
		li.InsertItem(len(config.Hosts), newitem)
	}
	// add segfault.net (free root server provider)
	li.InsertItem(len(config.Hosts), item{
		title: "segfault.net",
		desc:  "create free root server",
	})
	config.Hosts = append(config.Hosts, sshconf.Host{
		Name:    "segfault.net: free research server",
		Options: &safeorderedmap.SafeOrderedMap[string]{},
	})

	return li
}
