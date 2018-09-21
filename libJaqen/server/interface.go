package server

import (
	"sync"
	"time"
)

type Listener interface {
	//general
	//GetName will return the listeners 'name'. This will act as a prefix for the CLI if multiple listeners are defined.
	GetName() string
	//GetInfo will return the listener 'info' object. The 'info' Object should represent a high level of the current state of the listener (running state, options, name etc)
	GetInfo() ListenerInfo

	//options
	//SetOption will set the specified option.
	SetOption(string, string)
	//GetOption will return the specified option, or an error if the option does not exist.
	GetOption(string) (string, error)
	//GetOptions will return all valid options for the listener.
	GetOptions() func(string) []string

	//agent stuff
	//Kill will send a 'kill' message to the agent.
	Kill(string)
	//SendCommand will send the specified agent a command.
	SendCommand(string, string)
	//GenerateAgentFormats returns a slice of agent formats that can be generated.
	GenerateAgentFormats() []string
	//Generate will generate an agent of the specified format.
	Generate(string) []byte

	//runstate
	//Init should initialise zero values of the listener, and return the signal channels required to interact with the listener
	Init() (SignalChans, error)
	//Start will start the listener
	Start() error
	//Stop will stop the listener (but not kill any agents).
	Stop()
}

//Checkin defins the potential information that can be sent by agents to a checkin channel
type Checkin struct {
	UID          string
	ListenerName string
	AgentOS      string
	AgentType    string
	AgentArch    string
	CheckinTime  time.Time
}

//ListenerOptions is a concurrency safe map implementation, suitible for storing listener options
type ListenerOptions struct {
	options   map[string]string
	muOptions *sync.RWMutex
}

//New will return an empty initialised Listeneroptions object
func (l ListenerOptions) New() ListenerOptions {
	r := ListenerOptions{
		options:   make(map[string]string),
		muOptions: &sync.RWMutex{},
	}
	return r
}

//Set will set the value
func (l *ListenerOptions) Set(key, value string) {
	l.muOptions.Lock()
	defer l.muOptions.Unlock()
	l.options[key] = value
}

//Get returns the option value and 'true' if it exists, or an empty string and 'false' if it does not.
func (l *ListenerOptions) Get(k string) (string, bool) {
	l.muOptions.RLock()
	defer l.muOptions.RUnlock()
	if v, ok := l.options[k]; ok {
		return v, true
	}
	return "", false
}

//GetKeys returns a function that returns a slice of strings representing all options for the listener
func (l *ListenerOptions) GetKeys() func(string) []string {
	l.muOptions.RLock()
	defer l.muOptions.RUnlock()
	return func(line string) []string {
		a := make([]string, 0)
		for k := range l.options { //other options go here?
			a = append(a, k)
		}
		return a
	}
}

//CloneMap returns the options map defined internally (deep copy)
func (l *ListenerOptions) CloneMap() map[string]string {
	l.muOptions.RLock()
	defer l.muOptions.RUnlock()
	r := make(map[string]string)
	for x := range l.options {
		r[x] = l.options[x]
	}
	return r
}

//SignalChans contains the signal channels required in order to communicate with the listener.
type SignalChans struct {
	CheckinChan  chan Checkin
	ResponseChan chan Command
}

//ListenerInfo contains the running state of the listener.
type ListenerInfo struct {
	Name    string
	Running bool
	Options map[string]string
}
