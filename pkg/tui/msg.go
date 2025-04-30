package tui

type (
	ShowConfigMsg    struct{}
	ReloadConfigMsg  struct{}
	LivenessCheckMsg struct{}
	ExitOnConnMsg    struct{}
	tickMsg          struct{}
	AppMsg           struct {
		Text string
	}
	FilterTagMsg struct {
		Arg string
	}
)
