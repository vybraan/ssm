package tui

type SysCmd string

func (s SysCmd) String() string {
	return string(s)
}

const (
	ssh  SysCmd = "ssh"
	mosh SysCmd = "mosh"
)
