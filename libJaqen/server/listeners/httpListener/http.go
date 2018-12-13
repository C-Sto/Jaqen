package httpListener

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/c-sto/Jaqen/libJaqen/server"
)

type httpResponse struct {
	UID  string
	resp byte
}

type JaqenhttpListener struct {
	options          server.ListenerOptions
	httpResponseChan chan httpResponse
	CheckinChan      chan server.Checkin
	ResponseChan     chan server.Command

	agents          map[string]*agent
	aMut            *sync.RWMutex
	pendingCommands map[string][]string
	pcMut           *sync.RWMutex

	running bool
}

func (t *JaqenhttpListener) marshalResponses() {
	buffer := []byte{}
	for {
		//take responses from the local channel, and push them into the global jaqen response channel
		b := <-t.httpResponseChan
		buffer = append(buffer, b.resp)
		if string(b.resp) == "\n" {
			t.ResponseChan <- server.Command{
				UID:      b.UID,
				Response: buffer,
			}

			buffer = []byte{}
		}
	}
}

func (t *JaqenhttpListener) Init() (server.SignalChans, error) {
	t.options = server.ListenerOptions{}.New()
	t.pendingCommands = make(map[string][]string)
	t.pcMut = &sync.RWMutex{}
	t.aMut = &sync.RWMutex{}
	t.CheckinChan = make(chan server.Checkin, 10)
	t.ResponseChan = make(chan server.Command, 10)
	t.httpResponseChan = make(chan httpResponse, 1)

	//defualt options
	t.options.Set("port", "80")
	t.options.Set("host", "0.0.0.0")
	t.options.Set("hostname", "hostname.com")
	t.options.Set("exectime", "500")     //ms
	t.options.Set("checkintime", "1000") //ms

	t.agents = make(map[string]*agent)

	r := server.SignalChans{
		CheckinChan:  t.CheckinChan,
		ResponseChan: t.ResponseChan,
	}
	go t.marshalResponses()
	//	go t.listenCheckins()

	t.running = false

	return r, nil
}

func (t *JaqenhttpListener) listen() {
	//set up listener things
	for {
		//on checkin actions
		//conn, err := l.Accept()

		//push new agent to the checkin channel
		go func() {
			a := newAgent() //conn, t.httpResponseChan)
			t.aMut.Lock()
			t.agents[a.GetID()] = a
			t.aMut.Unlock()
			t.CheckinChan <- server.Checkin{
				UID:          a.GetID(),
				ListenerName: t.GetName(),
				CheckinTime:  time.Now(),
			}
		}()
	}
}

func newAgent() *agent { //con net.Conn, outChan chan httpResponse) *agent {
	//set up a new agent instance (decide how to talk to the agent etc)
	a := &agent{
		inChan: make(chan string),
		//outChan: outChan,
		//conn:    con,
		//reader:  reader,
		//writer:  writer,
	}

	a.Listen()

	return a
}

func (t JaqenhttpListener) GetName() string {
	return "http"
}

func (t *JaqenhttpListener) SetOption(k, v string) {
	t.options.Set(k, v)
}

func (t *JaqenhttpListener) GetOption(s string) (string, error) {
	if x, ok := t.options.Get(s); ok {
		return x, nil
	}
	return "", errors.New("Option not found")
}

func (t *JaqenhttpListener) GetOptions() func(string) []string {
	return t.options.GetKeys()
}

func (t *JaqenhttpListener) Kill(id string) {

}

func (t *JaqenhttpListener) SendCommand(agent, command string) {
	//ensure commands begin with a newline
	if string(command[len(command)-1]) != "\n" {
		command = command + "\n"
	}
	t.aMut.Lock()
	t.agents[agent].inChan <- command
	t.aMut.Unlock()
}

func (t *JaqenhttpListener) Start() error {
	p, _ := t.GetOption("port")
	fmt.Println("Starting http Listerner on port " + p)
	go t.listen()
	t.running = true
	return nil
}

func (t *JaqenhttpListener) Stop() {}

func (t JaqenhttpListener) GenerateAgentFormats() []string {
	return []string{"golang"} // soon.jpg ,"powershell", "bash"}
}

func (t JaqenhttpListener) Generate(s string) []byte {
	switch strings.ToLower(s) {
	//case "powershell":
	//	return []byte(t.genPowershellAgent())
	case "golang":
		return t.genGolangAgent()
		//case "bash":
		//	return []byte(t.genBashAgent())
	}
	return []byte{}
}

func (t JaqenhttpListener) GetInfo() server.ListenerInfo {
	//copy the options map to avoid bad stuff
	newMap := t.options.CloneMap()

	return server.ListenerInfo{
		Name:    t.GetName(),
		Running: t.running,
		Options: newMap,
	}
}
