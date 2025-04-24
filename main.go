package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lfaoro/ssm/pkg/sshconf"
	"github.com/lfaoro/ssm/pkg/tui"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

var BuildVersion = "v0.0.1-dev"
var BuildDate = "unset"
var BuildSHA = "unset"

var (
	filterArg string
)

func main() {
	appcmd := &cli.Command{
		Authors: []any{
			map[string]string{
				"name":  "Leonardo Faoro",
				"email": "ssm@leonardofaoro.com",
			},
		},
		Name:                   "ssm",
		EnableShellCompletion:  true,
		UseShortOptionHandling: true,
		Suggest:                true,

		Version: fmt.Sprintf("%s\nbuild date: %s\nbuild SHA: %s", BuildVersion, BuildDate, BuildSHA),
		ExtraInfo: func() map[string]string {
			return map[string]string{
				"Build version": BuildVersion,
				"Build date":    BuildDate,
				"Build sha":     BuildSHA,
			}
		},
		Usage:       "Secure Shell Manager",
		ArgsUsage:   "[filter]",
		Description: "SSM allows easy connection to SSH servers, hosts filtering, editing, tagging, command execution and file transfer.",

		Before: func(c context.Context, cmd *cli.Command) (context.Context, error) {
			_ = cmd
			return c, nil
		},

		Action: mainCmd,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "filter",
				Destination: &filterArg,
				Max:         1,
			},
		},

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "enable debug mode with verbose logging",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "exit",
				Aliases: []string{"e"},
				Usage:   "exit after connecting to a host",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "info",
				Aliases: []string{"i"},
				Usage:   "always show config keys",
				Value:   false,
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "custom config file path",
			},
		},

		Commands: []*cli.Command{
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
			fmt.Println("you found a bug#1: open an issue")
			os.Exit(1)
		}
		if m.ExitOnCmd {
			sshPath, err := exec.LookPath(m.ExtCmd)
			if err != nil {
				fmt.Printf("can't find `%s` cmd in your path: %v\n", m.ExtCmd, err)
				os.Exit(1)
			}
			fmt.Printf("ssm will exit and be replaced by %s\n", m.ExtCmd)
			err = syscall.Exec(sshPath, []string{"ssh", m.ExitHost}, os.Environ())
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}()
	if filterArg != "" {
		p.Send(tui.FilterMsg{
			Arg: fmt.Sprintf("#%s", filterArg),
		})
	}
	if cmd.Bool("exit") {
		p.Send(tui.ExitOnConnMsg{})
	}
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
