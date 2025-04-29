package tui

type AppMsg struct {
	Text string
}
type ShowConfigMsg struct{}
type ReloadConfigMsg struct{}
type LivenessCheckMsg struct{}
type ExitOnConnMsg struct{}
type FilterTagMsg struct {
	Arg string
}
