package tui

type State int

const (
	ListHosts State = iota
	RunCommand
)
