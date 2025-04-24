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
		Copyright:              "Leonardo Faoro (MIT)",
		UsageText:              "ssm [--options] [tag]\nexample: ssm --show --exit vpn\nexample: ssm -se vpn",

		Version: fmt.Sprintf(`%s
	 build date: %s
	 build SHA: %s`, BuildVersion, BuildDate, BuildSHA),
		ExtraInfo: func() map[string]string {
			return map[string]string{
				"Build version": BuildVersion,
				"Build date":    BuildDate,
				"Build sha":     BuildSHA,
			}
		},
		Usage:       "Secure Shell Manager",
		ArgsUsage:   "[tag]",
		Description: "SSM is an open source (MIT) SSH connection manager that helps engineers organize servers, connect, filter, tag, execute commands (soon), transfer files (soon), and much more from a simple terminal interface.",

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
				Usage:   "always show config",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "exit",
				Aliases: []string{"e"},
				Usage:   "exit after connection",
				Value:   false,
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
