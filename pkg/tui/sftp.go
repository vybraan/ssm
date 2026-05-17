package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	pathpkg "path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/lfaoro/ssm/pkg/sshconf"
	"github.com/pkg/sftp"
)

type paneSide int

const (
	localPane paneSide = iota
	remotePane
)

type sftpConnectMsg struct {
	cmd    *exec.Cmd
	stderr *bytes.Buffer
	client *sftp.Client
	root   string
	err    error
}

type sftpDirMsg struct {
	side  paneSide
	path  string
	items []list.Item
	err   error
}

type sftpTransferMsg struct {
	text          string
	err           error
	refreshLocal  bool
	refreshRemote bool
}

type fileItem struct {
	name     string
	path     string
	isDir    bool
	size     int64
	modTime  time.Time
	kind     string
	selected bool
}

func (f fileItem) Title() string {
	if f.isDir || f.name == ".." {
		return f.name
	}
	prefix := "  "
	if f.selected {
		prefix = "● "
	}
	return prefix + f.name
}

func (f fileItem) Description() string {
	if f.name == ".." {
		return "parent directory"
	}

	size := "--"
	if !f.isDir {
		size = humanSize(f.size)
	}
	stamp := "--"
	if !f.modTime.IsZero() {
		stamp = f.modTime.Format("2006-01-02 15:04")
	}
	return fmt.Sprintf("%-16s  %-8s  %s", stamp, size, f.kind)
}

func (f fileItem) FilterValue() string {
	return f.name
}

type filePane struct {
	title string
	cwd   string
	list  list.Model
}

type paneDelegate struct {
	list.DefaultDelegate
}

func (d paneDelegate) Render(w io.Writer, m list.Model, idx int, item list.Item) {
	if fi, ok := item.(fileItem); ok && fi.selected {
		d.Styles.NormalTitle = d.Styles.NormalTitle.Foreground(lg.Color("#FFA500"))
		d.Styles.NormalDesc = d.Styles.NormalDesc.Foreground(lg.Color("#CC8400"))
		d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(lg.Color("#FFA500"))
		d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(lg.Color("#CC8400"))
	}
	d.DefaultDelegate.Render(w, m, idx, item)
}

type sftpModel struct {
	previous   *Model
	host       sshconf.Host
	firstBoot  bool
	width      int
	height     int
	activePane paneSide
	local      filePane
	remote     filePane
	sshCmd     *exec.Cmd
	sshErr     *bytes.Buffer
	sftpClient *sftp.Client
	status     string
	selected   map[string]bool
	showHidden bool
}

// SftpModel wraps the base model in an SFTP browser sub-model.
func SftpModel(base tea.Model) tea.Model {
	previous, ok := base.(*Model)
	if !ok {
		panic("failed to cast tea.Model to Model")
	}

	i := previous.li.GlobalIndex()
	host := previous.config.Hosts[i]
	startDir, err := os.Getwd()
	if err != nil {
		startDir, _ = os.UserHomeDir()
	}

	m := &sftpModel{
		previous:   previous,
		host:       host,
		firstBoot:  true,
		activePane: localPane,
		local:      newFilePane("Local", startDir),
		remote:     newFilePane(host.Name, "."),
		status:     "Connecting...",
		selected:   map[string]bool{},
	}
	if items, err := loadLocalDir(startDir, false); err == nil {
		m.local.list.SetItems(items)
	}
	m.syncPaneSizes(previous.li.Width(), previous.li.Height())
	return m
}

func newFilePane(title, cwd string) filePane {
	dd := list.NewDefaultDelegate()
	dd.SetSpacing(0)
	dd.ShowDescription = true
	d := paneDelegate{dd}

	li := list.New([]list.Item{}, d, 0, 0)
	li.DisableQuitKeybindings()
	li.SetFilteringEnabled(true)
	li.SetShowHelp(false)
	li.SetShowPagination(false)
	li.SetShowStatusBar(false)
	li.SetShowFilter(true)

	return filePane{
		title: title,
		cwd:   cwd,
		list:  li,
	}
}

func (s *sftpModel) Init() tea.Cmd {
	return nil
}

func (s *sftpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if s.firstBoot {
		s.firstBoot = false
		cmds = append(cmds,
			connectRemoteCmd(s.host, s.previous.config.GetPath()),
		)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.syncPaneSizes(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyEsc:
			if s.activeList().FilterState() == list.Filtering {
				s.activeList().ResetFilter()
				break
			}
			s.close()
			return s.previous, nil
		case tea.KeyTab:
			s.toggleFocus()
		case tea.KeyLeft:
			s.activePane = localPane
		case tea.KeyRight:
			s.activePane = remotePane
		case tea.KeyEnter:
			if cmd := s.handleEnter(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		case tea.KeySpace:
			if s.activeList().FilterState() != list.Filtering {
				s.toggleSelection()
			}
		case '.':
			if s.activeList().FilterState() != list.Filtering {
				s.showHidden = !s.showHidden
				cmds = append(cmds, s.reloadActiveDir())
			}
		}
	case sftpConnectMsg:
		if msg.err != nil {
			s.status = msg.err.Error()
			return s, tea.Batch(cmds...)
		}
		s.sshCmd = msg.cmd
		s.sshErr = msg.stderr
		s.sftpClient = msg.client
		s.remote.cwd = msg.root
		s.status = fmt.Sprintf("Connected to %s", s.host.Name)
		cmds = append(cmds, loadRemoteDirCmd(s.sftpClient, s.remote.cwd, false))
	case sftpDirMsg:
		if msg.err != nil {
			s.status = msg.err.Error()
			break
		}
		clear(s.selected)
		if msg.side == localPane {
			s.local.cwd = msg.path
			s.local.list.SetItems(msg.items)
			s.local.list.SetItems(s.markSelections(s.local.list.Items()))
		} else {
			s.remote.cwd = msg.path
			s.remote.list.SetItems(msg.items)
			s.remote.list.SetItems(s.markSelections(s.remote.list.Items()))
		}
	case sftpTransferMsg:
		if msg.err != nil {
			s.status = msg.err.Error()
		} else {
			s.status = msg.text
		}
		if msg.refreshLocal {
			cmds = append(cmds, loadLocalDirCmd(s.local.cwd, s.showHidden))
		}
		if msg.refreshRemote && s.sftpClient != nil {
			cmds = append(cmds, loadRemoteDirCmd(s.sftpClient, s.remote.cwd, s.showHidden))
		}
	}

	var cmd tea.Cmd
	if s.activePane == localPane {
		s.local.list, cmd = s.local.list.Update(msg)
	} else {
		s.remote.list, cmd = s.remote.list.Update(msg)
	}
	cmds = append(cmds, cmd)

	return s, tea.Batch(cmds...)
}

func (s *sftpModel) handleEnter() tea.Cmd {
	var pane *filePane
	var navigate func(string) tea.Cmd
	var transfer func(*sftp.Client, string, string) tea.Cmd
	if s.activePane == localPane {
		pane = &s.local
		navigate = func(p string) tea.Cmd { return loadLocalDirCmd(p, s.showHidden) }
		transfer = uploadFileCmd
	} else {
		pane = &s.remote
		navigate = func(p string) tea.Cmd { return loadRemoteDirCmd(s.sftpClient, p, s.showHidden) }
		transfer = downloadFileCmd
	}

	item, ok := pane.list.SelectedItem().(fileItem)
	if !ok {
		return nil
	}

	if len(s.selected) > 0 {
		return s.batchTransfer(transfer)
	}

	if item.isDir {
		return navigate(item.path)
	}

	if s.sftpClient == nil {
		return transferMsgCmd("", fmt.Errorf("not connected to remote host"), false, false)
	}

	return s.transferOne(item, transfer)
}

func (s *sftpModel) transferOne(item fileItem, transfer func(*sftp.Client, string, string) tea.Cmd) tea.Cmd {
	var src, dst string
	if s.activePane == localPane {
		src, dst = item.path, pathpkg.Join(s.remote.cwd, filepath.Base(item.path))
	} else {
		src, dst = item.path, filepath.Join(s.local.cwd, pathpkg.Base(item.path))
	}
	return transfer(s.sftpClient, src, dst)
}

func (s *sftpModel) batchTransfer(transfer func(*sftp.Client, string, string) tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		var paths []string
		var pane *filePane
		var dstPath func(string) string
		if s.activePane == localPane {
			pane = &s.local
			dstPath = func(p string) string { return pathpkg.Join(s.remote.cwd, filepath.Base(p)) }
		} else {
			pane = &s.remote
			dstPath = func(p string) string { return filepath.Join(s.local.cwd, pathpkg.Base(p)) }
		}

		for _, it := range pane.list.Items() {
			if fi, ok := it.(fileItem); ok && s.selected[fi.path] && !fi.isDir {
				paths = append(paths, fi.path)
			}
		}
		if len(paths) == 0 {
			return sftpTransferMsg{text: "no files selected", refreshLocal: false, refreshRemote: false}
		}

		var errs []string
		for _, p := range paths {
			msg := transfer(s.sftpClient, p, dstPath(p))
			result := msg().(sftpTransferMsg)
			if result.err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", filepath.Base(p), result.err))
			}
		}

		text := fmt.Sprintf("transferred %d/%d files", len(paths)-len(errs), len(paths))
		if len(errs) > 0 {
			return sftpTransferMsg{
				text: text, err: fmt.Errorf("%s", strings.Join(errs, "; ")),
				refreshLocal: true, refreshRemote: s.sftpClient != nil,
			}
		}
		return sftpTransferMsg{text: text, refreshLocal: true, refreshRemote: s.sftpClient != nil}
	}
}

func (s *sftpModel) toggleSelection() {
	var pane *filePane
	if s.activePane == localPane {
		pane = &s.local
	} else {
		pane = &s.remote
	}
	item, ok := pane.list.SelectedItem().(fileItem)
	if !ok || item.name == ".." {
		return
	}
	if s.selected[item.path] {
		delete(s.selected, item.path)
	} else {
		s.selected[item.path] = true
	}
	pane.list.SetItems(s.markSelections(pane.list.Items()))
}

func (s *sftpModel) markSelections(items []list.Item) []list.Item {
	result := make([]list.Item, len(items))
	for i, it := range items {
		fi, ok := it.(fileItem)
		if !ok {
			result[i] = it
			continue
		}
		if s.selected[fi.path] {
			fi.selected = true
		} else {
			fi.selected = false
		}
		result[i] = fi
	}
	return result
}

func (s *sftpModel) toggleFocus() {
	if s.activePane == localPane {
		s.activePane = remotePane
		return
	}
	s.activePane = localPane
}

func (s *sftpModel) activeList() *list.Model {
	if s.activePane == localPane {
		return &s.local.list
	}
	return &s.remote.list
}

func (s *sftpModel) reloadActiveDir() tea.Cmd {
	if s.activePane == localPane {
		return loadLocalDirCmd(s.local.cwd, s.showHidden)
	}
	if s.sftpClient != nil {
		return loadRemoteDirCmd(s.sftpClient, s.remote.cwd, s.showHidden)
	}
	return nil
}

func (s *sftpModel) close() {
	if s.sftpClient != nil {
		_ = s.sftpClient.Close()
		s.sftpClient = nil
	}
	if s.sshCmd != nil && s.sshCmd.Process != nil {
		_ = s.sshCmd.Process.Kill()
		_, _ = s.sshCmd.Process.Wait()
		s.sshCmd = nil
	}
}

func (s *sftpModel) syncPaneSizes(width, height int) {
	if width <= 0 {
		width = s.previous.li.Width()
	}
	if height <= 0 {
		height = s.previous.li.Height()
	}
	s.width = width
	s.height = height

	paneWidth := max(20, (width/2)-2)
	paneHeight := max(10, height-6)

	s.local.list.SetSize(paneWidth, paneHeight)
	s.remote.list.SetSize(paneWidth, paneHeight)
}

func (s *sftpModel) View() tea.View {
	bar := fmt.Sprintf("sftp  %s", s.status)
	if n := len(s.selected); n > 0 {
		bar = fmt.Sprintf("sftp ● %d  %s", n, s.status)
	}
	v := lg.JoinVertical(
		lg.Left,
		lg.NewStyle().Foreground(lg.Color("8")).Render(bar),
		"",
		lg.JoinHorizontal(lg.Top,
			s.renderPane(s.local, s.activePane == localPane),
			lg.NewStyle().Foreground(lg.Color("8")).Render("│"),
			s.renderPane(s.remote, s.activePane == remotePane),
		),
	)
	view := tea.NewView(v)
	view.AltScreen = true
	return view
}

func (s *sftpModel) renderPane(p filePane, focused bool) string {
	headerStyle := lg.NewStyle().
		Bold(focused).
		Foreground(lg.Color("8"))
	if focused {
		headerStyle = headerStyle.Foreground(lg.Color(s.previous.theme.selectedTitleColor))
	}

	header := headerStyle.Render(fmt.Sprintf("%s  %s", p.title, p.cwd))
	underline := lg.NewStyle().
		Foreground(lg.Color("8")).
		Render(strings.Repeat("─", max(1, p.list.Width())))
	if focused {
		underline = lg.NewStyle().
			Foreground(lg.Color(s.previous.theme.selectedBorderColor)).
			Render(strings.Repeat("─", max(1, p.list.Width())))
	}

	body := lg.NewStyle().
		Padding(0, 1, 0, 0).
		Width(p.list.Width()).
		Render(p.list.View())

	return lg.NewStyle().
		Width(p.list.Width() + 1).
		Render(header + "\n" + underline + "\n" + body)
}

func loadLocalDirCmd(path string, showHidden bool) tea.Cmd {
	return func() tea.Msg {
		items, err := loadLocalDir(path, showHidden)
		return sftpDirMsg{
			side:  localPane,
			path:  path,
			items: items,
			err:   err,
		}
	}
}

func loadRemoteDirCmd(client *sftp.Client, path string, showHidden bool) tea.Cmd {
	return func() tea.Msg {
		items, err := loadRemoteDir(client, path, showHidden)
		return sftpDirMsg{
			side:  remotePane,
			path:  path,
			items: items,
			err:   err,
		}
	}
}

func connectRemoteCmd(host sshconf.Host, configPath string) tea.Cmd {
	return func() tea.Msg {
		cmd, stderr, client, root, err := connectSFTP(host, configPath)
		return sftpConnectMsg{
			cmd:    cmd,
			stderr: stderr,
			client: client,
			root:   root,
			err:    err,
		}
	}
}

func uploadFileCmd(client *sftp.Client, src, dst string) tea.Cmd {
	return func() tea.Msg {
		srcFile, err := os.Open(src) //nolint:gosec
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		defer func() { _ = srcFile.Close() }()

		dstFile, err := client.Create(dst)
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		defer func() { _ = dstFile.Close() }()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return sftpTransferMsg{err: err}
		}

		return sftpTransferMsg{
			text:          fmt.Sprintf("uploaded %s", filepath.Base(src)),
			refreshRemote: true,
		}
	}
}

func downloadFileCmd(client *sftp.Client, src, dst string) tea.Cmd {
	return func() tea.Msg {
		srcFile, err := client.Open(src)
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		defer func() { _ = srcFile.Close() }()

		dstFile, err := os.Create(dst) //nolint:gosec
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		defer func() { _ = dstFile.Close() }()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return sftpTransferMsg{err: err}
		}

		return sftpTransferMsg{
			text:         fmt.Sprintf("downloaded %s", pathpkg.Base(src)),
			refreshLocal: true,
		}
	}
}

func transferMsgCmd(text string, err error, refreshLocal, refreshRemote bool) tea.Cmd {
	return func() tea.Msg {
		return sftpTransferMsg{
			text:          text,
			err:           err,
			refreshLocal:  refreshLocal,
			refreshRemote: refreshRemote,
		}
	}
}

func loadLocalDir(path string, showHidden bool) ([]list.Item, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	items := make([]list.Item, 0, len(entries)+1)
	parent := filepath.Dir(path)
	if parent != path {
		items = append(items, fileItem{
			name:  "..",
			path:  parent,
			isDir: true,
			kind:  "dir",
		})
	}

	for _, entry := range entries {
		if !showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		items = append(items, fileItem{
			name:    entry.Name(),
			path:    filepath.Join(path, entry.Name()),
			isDir:   entry.IsDir(),
			size:    info.Size(),
			modTime: info.ModTime(),
			kind:    fileKind(entry.Name(), entry.IsDir()),
		})
	}
	return items, nil
}

func loadRemoteDir(client *sftp.Client, path string, showHidden bool) ([]list.Item, error) {
	entries, err := client.ReadDir(path)
	if err != nil {
		return nil, err
	}

	items := make([]list.Item, 0, len(entries)+1)
	parent := pathpkg.Dir(path)
	if parent != path {
		items = append(items, fileItem{
			name:  "..",
			path:  parent,
			isDir: true,
			kind:  "dir",
		})
	}

	for _, entry := range entries {
		if !showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		items = append(items, fileItem{
			name:    entry.Name(),
			path:    pathpkg.Join(path, entry.Name()),
			isDir:   entry.IsDir(),
			size:    entry.Size(),
			modTime: entry.ModTime(),
			kind:    fileKind(entry.Name(), entry.IsDir()),
		})
	}
	return items, nil
}

func connectSFTP(host sshconf.Host, configPath string) (*exec.Cmd, *bytes.Buffer, *sftp.Client, string, error) {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("ssh not found in PATH: %w", err)
	}

	cmd := exec.Command(sshPath, "-F", configPath, host.Name, "-s", "sftp") //nolint:gosec
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, "", err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, "", err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, nil, "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}

	client, err := sftp.NewClientPipe(stdout, stdin)
	if err != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, nil, nil, "", fmt.Errorf("%w: %s", err, msg)
		}
		return nil, nil, nil, "", err
	}

	root, err := client.Getwd()
	if err != nil || root == "" {
		root = "."
	}

	return cmd, stderr, client, root, nil
}

func fileKind(name string, isDir bool) string {
	if isDir {
		return "dir"
	}
	ext := strings.TrimPrefix(filepath.Ext(name), ".")
	if ext == "" {
		return "file"
	}
	return ext
}

func humanSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	value := float64(size) / float64(div)
	return strconv.FormatFloat(value, 'f', 1, 64) + string("KMGTPE"[exp]) + "B"
}
