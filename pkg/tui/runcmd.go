package tui

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/v2/spinner"
	"github.com/charmbracelet/bubbles/v2/textinput"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

func RunCmdModel(base tea.Model) tea.Model {
	previousModel, ok := base.(*Model)
	if !ok {
		panic("failed to cast tea.Model to Model")
	}

	cmdInput := textinput.New()
	vp := viewport.New()

	cmdInput.Placeholder = "enter command"
	cmdInput.Prompt = "> "
	cmdInput.CharLimit = 256
	cmdInput.Focus()
	cmdInput.VirtualCursor = true

	// using double main model viewport width because it use half of screenwidth
	cmdInput.SetWidth(previousModel.vp.Width()*2 - 3)
	vp.SetWidth(previousModel.vp.Width() * 2)

	vp.SetHeight(previousModel.vp.Height() - lipgloss.Height(cmdInput.View()) - 2) // - 2 to accommodate the bar, since we can't get the Height
	vp.SetContent("(no output) ...")

	s := spinner.New()
	s.Spinner = spinner.Dot
	return &cmdModel{
		previousModel: base,
		viewport:      vp,
		input:         cmdInput,
		ready:         false,
		running:       false,
		commands:      []string{},
		spinner:       s,
	}
}

type cmdResultMsg struct {
	output string
	err    error
}

type cmdModel struct {
	commands      []string
	previousModel tea.Model
	viewport      viewport.Model
	input         textinput.Model
	ready         bool
	running       bool
	spinner       spinner.Model
}

func (m *cmdModel) Init() tea.Cmd {
	return nil
}

func (m *cmdModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	if m.running {
		cmds = append(cmds, m.spinner.Tick)
	}

	if !m.ready {
		m.ready = true
		cmds = append(cmds, textinput.Blink)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.handleWindowSize(msg)

	case cmdResultMsg:
		m.handleCommandResult(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m *cmdModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.previousModel, nil

	case "enter":
		command := strings.TrimSpace(m.input.Value())
		if command == "" {
			return m, nil
		}

		m.input.SetValue("")
		m.commands = append(m.commands, "$ "+command)
		m.viewport.SetContent(strings.Join(m.commands, "\n"))
		m.viewport.GotoBottom()

		m.input.Blur()
		m.running = true

		return m, runCommand(m, command)
	}

	return m, nil
}

func (m *cmdModel) handleWindowSize(msg tea.WindowSizeMsg) {
	m.input.SetWidth(msg.Width - 3)
	m.viewport.SetWidth(msg.Width)
	m.viewport.SetHeight(msg.Height)
}

func (m *cmdModel) handleCommandResult(msg cmdResultMsg) {
	if msg.err != nil {
		errorMsg := msg.err.Error() + "\n" + msg.output
		m.commands = append(m.commands, errorMsg)
		m.viewport.SetContent(strings.Join(m.commands, "\n"))
		m.viewport.GotoBottom()
	} else {
		m.commands = append(m.commands, msg.output)
		m.viewport.SetContent(strings.Join(m.commands, "\n"))
		m.viewport.GotoBottom()
	}

	m.running = false
	m.input.Focus()
}

func (m cmdModel) View() string {

	var builder strings.Builder

	builder.WriteString(m.Bar() + "\n\n")
	if m.running {

		builder.WriteString(m.spinner.View() + " " + m.input.View() + "\n\n")
	} else {

		builder.WriteString(m.input.View() + "\n\n")
	}
	builder.WriteString(m.viewport.View())

	return builder.String()
}
func (m cmdModel) Bar() string {
	pm, ok := m.previousModel.(*Model)
	if !ok {
		return "Invalid model"
	}

	selectedItem, ok := pm.li.SelectedItem().(item)
	if !ok {
		return renderPrimaryBar("No host selected", pm.theme.selectedTitleColor)
	}

	windowName := renderPrimaryBar("Run Command", pm.theme.selectedTitleColor)
	status := renderPrimaryBar("SSM", pm.theme.selectedTitleColor)

	availableWidth := m.viewport.Width() - lipgloss.Width(windowName) - lipgloss.Width(status)
	host := renderSecondaryBar(selectedItem.Description(), availableWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, windowName, host, status)
}

func renderPrimaryBar(content string, bgColor string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color(bgColor)).
		Padding(0, 1).
		Render(content)
}

func renderSecondaryBar(content string, width int) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#343433")).
		Padding(0, 1).
		Width(width).
		Render(content)
}

func runCommand(m *cmdModel, command string) tea.Cmd {
	return func() tea.Msg {
		prev, ok := m.previousModel.(*Model)
		if !ok {
			return cmdResultMsg{output: "", err: fmt.Errorf("invalid previous model")}
		}

		selected, ok := prev.li.SelectedItem().(item)
		if !ok {
			return cmdResultMsg{output: "", err: fmt.Errorf("no selected host")}
		}

		// ssh command args to force use of keys
		args := []string{
			"-T",
			"-F", prev.config.GetPath(),
			// "-o", "PreferredAuthentications=publickey",
			// "-o", "PasswordAuthentication=no",
			selected.title,
			command,
		}

		cmd := exec.Command(prev.Cmd.String(), args...)

		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()

		return cmdResultMsg{output: out.String(), err: err}
	}
}
