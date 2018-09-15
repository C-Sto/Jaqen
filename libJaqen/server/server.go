package server

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/fatih/color"
)

type OutType int

const (
	//CHECKIN - checkin output
	CHECKIN OutType = iota
	//RESPONSE - response output
	RESPONSE
	//INFO - informational output
	INFO
	//ERR - error output
	ERR
)

type Output struct {
	Val   string
	Type  OutType
	Level int
}

type JaqenServer struct {
	listeners map[string]Listener
	lMutex    *sync.RWMutex

	responseChans chan Command
	checkinChans  chan Checkin

	activeAgents map[string]Agent
	aMut         *sync.RWMutex
	outputChan   chan Output
}

type Command struct {
	UID      string
	Cmd      string
	Response []byte
}

type Agent struct {
	UID         string
	Listener    string
	Commands    map[string]Command
	LastCheckin time.Time
}

func (j *JaqenServer) GetOutputChan() chan Output {
	return j.outputChan
}

func (j *JaqenServer) Init() error {
	j.lMutex = &sync.RWMutex{}
	j.aMut = &sync.RWMutex{}
	j.listeners = make(map[string]Listener)
	j.activeAgents = make(map[string]Agent)
	j.checkinChans = make(chan Checkin)
	j.responseChans = make(chan Command)
	j.outputChan = make(chan Output, 10)
	go j.listenCheckins()
	go j.listenResponses()
	return nil
}

func (j *JaqenServer) addAgent(s string, a Agent) {
	j.aMut.Lock()
	defer j.aMut.Unlock()

	j.activeAgents[s] = a
}

func (j *JaqenServer) GetAgents() func(string) []string {
	j.aMut.RLock()
	defer j.aMut.RUnlock()

	return func(line string) []string {
		a := make([]string, 0)
		for k, _ := range j.activeAgents { //other options go here?
			a = append(a, k)
		}
		return a
	}

}

//exec
func (j *JaqenServer) AgentExecute(agent, command string) {
	//get agent
	j.aMut.Lock()
	defer j.aMut.Unlock()

	a := j.activeAgents[agent]

	//get listener
	j.lMutex.Lock()
	defer j.lMutex.Unlock()

	l := j.listeners[a.Listener]

	l.SendCommand(a.UID, command)

}

//show output
func (j *JaqenServer) AgentGetOutput(agent, command string) {

}

//kill
func (j *JaqenServer) AgentKill(agent string) {
	j.aMut.Lock()
	defer j.aMut.Unlock()

	l := j.activeAgents[agent].Listener
	j.listeners[l].Kill(agent)
	delete(j.activeAgents, agent)
}

func (j *JaqenServer) addCheckinChan(c <-chan Checkin) {
	go func(cc <-chan Checkin) {
		for s := range cc {
			j.checkinChans <- s
		}
	}(c)
}

func (j *JaqenServer) addResponseChan(c <-chan Command) {
	go func(cc <-chan Command) {
		for s := range cc {
			j.responseChans <- s
		}
	}(c)
}

func (j *JaqenServer) listenCheckins() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case c := <-j.checkinChans:
			j.aMut.Lock()
			if _, ok := j.activeAgents[c.UID]; ok {
				x := j.activeAgents[c.UID]
				x.LastCheckin = c.CheckinTime
				j.activeAgents[c.UID] = x
				j.aMut.Unlock()
				continue
			}
			j.aMut.Unlock()
			j.outputChan <- Output{
				Val: fmt.Sprintf("\n%s[%s]: %s", color.New(color.FgYellow).SprintfFunc()("Agent checkin"), c.ListenerName, color.New(color.Bold).SprintfFunc()(c.UID)),
			}
			a := Agent{
				UID:      c.UID,
				Commands: make(map[string]Command),
				Listener: c.ListenerName,
			}
			j.addAgent(c.UID, a)
		case <-ticker.C:
			agents := j.GetAgents()("")
			for _, x := range agents {
				a, r := j.getAgent(x)
				if !r {
					continue
				}

				//todo: set checkin timeout to listener dependent
				if !a.LastCheckin.IsZero() && time.Since(a.LastCheckin) > (time.Duration(30)*time.Second) {
					j.removeAgent(x)
				}
			}
		}
	}
}

func (j *JaqenServer) removeAgent(s string) {
	j.aMut.Lock()
	defer j.aMut.Unlock()
	fmt.Println("Removing ", s, "due to timeout")
	delete(j.activeAgents, s)
}

func (j *JaqenServer) getAgent(s string) (Agent, bool) {
	j.aMut.RLock()
	defer j.aMut.RUnlock()
	a, ok := j.activeAgents[s]
	return a, ok
}

func (j *JaqenServer) listenResponses() {
	for {
		select {
		case c := <-j.responseChans:
			j.addResponse(c)
		}
	}
}

func (j *JaqenServer) addResponse(c Command) {
	//gzzzzz need to keep map of responses so we can return later
	r := c.Response
	if string(c.Response[len(c.Response)-1]) == "\n" {
		r = r[:len(c.Response)-1]

	}
	j.outputChan <- Output{
		Val: fmt.Sprintf("%s[%s] %s: %s", color.New(color.FgGreen).SprintfFunc()("Response"), c.UID, c.Cmd, string(r)),
	}
}

func (j *JaqenServer) GetListeners() []ListenerInfo {
	j.lMutex.RLock()
	defer j.lMutex.RUnlock()

	r := []ListenerInfo{}
	for k, _ := range j.listeners {
		i := j.listeners[k].GetInfo()
		i.Name = k
		r = append(r, i)
	}
	return r
}

func (j *JaqenServer) StopListener(k string) {
	j.lMutex.Lock()
	defer j.lMutex.Unlock()
	j.listeners[k].Stop()
}

func (j *JaqenServer) RemoveListener(k string) {
	j.lMutex.Lock()
	defer j.lMutex.Unlock()
	if j.listeners[k].GetInfo().Running {
		j.listeners[k].Stop()
	}
	delete(j.listeners, k)
}

func (j *JaqenServer) AddListener(l Listener) (string, error) {
	j.lMutex.Lock()
	defer j.lMutex.Unlock()
	chans, e := l.Init()

	if e != nil {
		return "", e
	}

	j.addCheckinChan(chans.CheckinChan)
	j.addResponseChan(chans.ResponseChan)

	//check for existing listener of same type
	name := l.GetName()
	for {
		if _, ok := j.listeners[name]; ok {
			//already have listener, add unique suffix
			name = name + "-" + RandStringRunes(2)
		} else {
			break
		}
	}
	j.listeners[name] = l

	return name, nil
}

func (j *JaqenServer) SetOption(listener, option, value string) {
	j.lMutex.Lock()
	defer j.lMutex.Unlock()
	j.listeners[listener].SetOption(option, value)
}

func (j *JaqenServer) GetOption(listener, option string) (string, error) {
	j.lMutex.RLock()
	defer j.lMutex.RUnlock()
	o, e := j.listeners[listener].GetOption(option)
	return o, e
}

func (j *JaqenServer) GetOptions(listener string) (func(string) []string, error) {
	j.lMutex.RLock()
	defer j.lMutex.RUnlock()
	return j.listeners[listener].GetOptions(), nil
}

func (j *JaqenServer) Start(listener string) error {
	j.lMutex.Lock()
	defer j.lMutex.Unlock()
	return j.listeners[listener].Start()
}

func (j *JaqenServer) GenerateAgent(l, f string) []byte {

	return j.listeners[l].Generate(f)
}

func (j *JaqenServer) GetAgentFormats(l string) []string {
	return j.listeners[l].GenerateAgentFormats()
}

func RandStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano()) //probably a bad idea, whatever
	letterRunes := []byte("abcdefghijklmnopqrstuvwxyz1234567890")
	b := make([]byte, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
