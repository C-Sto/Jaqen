package server

import "time"

type Checkin struct {
	UID          string
	ListenerName string
	AgentOS      string
	AgentType    string
	AgentArch    string
	CheckinTime  time.Time
}

type CommandSignal struct {
}

type SignalChans struct {
	CheckinChan  chan Checkin
	ResponseChan chan Command
}

type ListenerInfo struct {
	Name    string
	Running bool
	Options map[string]string
}

type Listener interface {
	//general
	GetName() string
	//This should be filled with the listener state so you can see it good in the CLI. Running state, options, and name (name will be set by server, don't stress about that too much)
	GetInfo() ListenerInfo

	//options
	SetOption(string, string)
	GetOption(string) (string, error)
	GetOptions() func(string) []string

	//agent stuff
	Kill(string)
	SendCommand(string, string)
	GenerateAgentFormats() []string
	Generate(string) []byte

	//runstate
	Init() (SignalChans, error)
	Start() error
	Stop()
}
