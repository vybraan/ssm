// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"fmt"
	"net/mail"
	"os"
	"os/exec"
	"sync"
	"syscall"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/google/go-github/github"
	"github.com/lfaoro/ssm/pkg/sshconf"
	"github.com/lfaoro/ssm/pkg/tui"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

var BuildVersion = "0.0.0-dev"
var BuildDate = "unset"
var BuildSHA = "unset"

// cli arguments
var (
	filterTag string
)

func main() {
	appcmd := &cli.Command{
		Name: "ssm",
		Authors: []any{
			&mail.Address{
				Name:    "Leonardo Faoro",
				Address: "ssm@leonardofaoro.com",
			},
		},
		EnableShellCompletion:  true,
		UseShortOptionHandling: true,
		Suggest:                true,
		Copyright:              "(c) Leonardo Faoro (MIT)",
		Usage:                  "Secure Shell Manager",
		UsageText:              "ssm [--options] [tag]\nexample: ssm --show --exit vpn\nexample: ssm -se vpn",
		ArgsUsage:              "[tag]",
		Description:            "SSM is an open source (MIT) SSH connection manager that helps engineers organize servers, connect, filter, tag, execute commands (soon), transfer files (soon), and much more from a simple terminal interface.",

		Version: BuildVersion,
		ExtraInfo: func() map[string]string {
			return map[string]string{
				"Build version": BuildVersion,
				"Build date":    BuildDate,
				"Build sha":     BuildSHA,
			}
		},

		Before: func(c context.Context, cmd *cli.Command) (context.Context, error) {
			_ = cmd
			return c, nil
		},

		Action: mainCmd,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "tag",
				UsageText:   "comma separated arguments for filtering #tag: hosts",
				Destination: &filterTag,
			},
		},

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "show",
				Aliases: []string{"s"},
				Usage:   "always show config params",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "exit",
				Aliases: []string{"e"},
				Usage:   "exit after connection",
				Value:   false,
			},
			&cli.BoolFlag{
				// TODO: not implemented
				Name:    "ping",
				Aliases: []string{"p"},
				Usage:   "ping hosts and show liveness",
				Value:   false,
				Hidden:  true,
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "custom config file path",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "enable debug mode with verbose logging",
				Value:   false,
			},
		},

		Commands: []*cli.Command{
			generateCmd,
			testCmd,
		},
	}

	err := appcmd.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func mainCmd(_ context.Context, cmd *cli.Command) error {
	debug := cmd.Bool("debug")
	if debug {
		for k, v := range cmd.ExtraInfo() {
			fmt.Println(k, v)
		}
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("not an interactive terminal :(")
	}

	var err error
	var config *sshconf.Config
	configFlag := cmd.String("config")
	if configFlag != "" {
		config, err = sshconf.ParsePath(configFlag)
		if err != nil {
			return err
		}
	} else {
		config, err = sshconf.Parse()
		if err != nil {
			return err
		}
	}

	m := tui.NewModel(config, debug)
	p := tea.NewProgram(
		m,
		tea.WithOutput(os.Stderr))
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		final, err := p.Run()
		if err != nil {
			e := fmt.Errorf("failed to run %v: %w", cmd.Name, err)
			fmt.Println(e)
			os.Exit(1)
		}
		m, ok := final.(*tui.Model)
		if !ok {
			fmt.Println("you found bug#1: open an issue")
			os.Exit(1)
		}
		if m.ExitOnCmd && m.ExitHost != "" {
			sshPath, err := exec.LookPath(m.Cmd.String())
			if err != nil {
				fmt.Printf("can't find `%s` cmd in your path: %v\n", m.Cmd, err)
				os.Exit(1)
			}
			err = syscall.Exec(sshPath, []string{"ssh", "-F", config.GetPath(), m.ExitHost}, os.Environ())
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}()
	if filterTag != "" {
		p.Send(tui.FilterTagMsg{
			Arg: fmt.Sprintf("#%s", filterTag),
		})
	}
	if cmd.Bool("exit") {
		p.Send(tui.ExitOnConnMsg{})
	}
	if cmd.Bool("show") {
		p.Send(tui.ShowConfigMsg{})
	}
	if cmd.Bool("ping") {
		p.Send(tui.LivenessCheckMsg{})
	}
	// inform user when new version is available
	go func() {
		tag, err := latestTag()
		if err != nil {
			if cmd.Bool("debug") {
				p.Send(tui.AppMsg{Text: fmt.Sprintf("%s", err)})
			}
			return
		}
		if tag != cmd.Version {
			msg := fmt.Sprintf("%s: new version %s is available", cmd.Version, tag)
			p.Send(tui.AppMsg{Text: msg})
		}
	}()
	wg.Wait()
	return nil
}

var testCmd = &cli.Command{
	Name:   "test",
	Action: testAction,
	Hidden: true,
}
var testAction = func(_ context.Context, cmd *cli.Command) error {
	return nil
}

var generateCmd = &cli.Command{
	Name:    "generate",
	Aliases: []string{"gen"},
	Action:  generateAction,
	Hidden:  true,
}
var generateAction = func(_ context.Context, cmd *cli.Command) error {
	return nil
}

func latestTag() (string, error) {
	client := github.NewClient(nil)
	owner := "lfaoro"
	repo := "ssm"

	tags, _, err := client.Repositories.ListTags(context.Background(), owner, repo, &github.ListOptions{PerPage: 1})
	if err != nil {
		return "", fmt.Errorf("failed to list tags: %v", err)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found in the repository")
	}

	return *tags[0].Name, nil
}
