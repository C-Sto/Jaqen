package tcpListener

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/c-sto/Jaqen/libJaqen/server"
)

func newAgent(con net.Conn, outChan chan tcpResponse) *agent {
	writer := bufio.NewWriter(con)
	reader := bufio.NewReader(con)

	a := &agent{
		inChan:  make(chan string),
		outChan: outChan,
		conn:    con,
		reader:  reader,
		writer:  writer,
	}

	a.Listen()

	return a
}

type tcpResponse struct {
	UID  string
	resp byte
}

type JaqenTCPListener struct {
	options    map[string]string
	optionsMut *sync.RWMutex
	//dnsCommandChan  chan dnsresponse
	//	dnsCheckinChan  chan dnsresponse
	//	dnsResponseChan chan dnsresponse
	tcpResponseChan chan tcpResponse //byte

	CommandChan  chan server.CommandSignal
	CheckinChan  chan server.Checkin
	ResponseChan chan server.Command

	agents map[string]*agent

	pendingCommands map[string][]string
	pcMut           *sync.RWMutex

	running bool
}

func (t *JaqenTCPListener) marshalResponses() {
	buffer := []byte{}
	for {
		b := <-t.tcpResponseChan
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

func (t *JaqenTCPListener) Init() (server.SignalChans, error) {
	x := new(sync.RWMutex)
	t.optionsMut = x
	t.options = make(map[string]string)
	t.pendingCommands = make(map[string][]string)
	t.pcMut = &sync.RWMutex{}
	//t.dnsCheckinChan = make(chan dnsresponse, 200)
	//d.dnsCommandChan = make(chan dnsresponse, 200)
	t.CheckinChan = make(chan server.Checkin, 10)
	t.CommandChan = make(chan server.CommandSignal, 10)
	t.ResponseChan = make(chan server.Command, 10)
	t.tcpResponseChan = make(chan tcpResponse, 1)
	t.options["port"] = ""
	t.options["ip"] = ""

	t.agents = make(map[string]*agent)

	r := server.SignalChans{
		CheckinChan:  t.CheckinChan,
		ResponseChan: t.ResponseChan,
	}
	go t.marshalResponses()
	go t.listenCheckins()

	t.running = false

	return r, nil
}

func (t *JaqenTCPListener) listen() {
	l, err := net.Listen("tcp", ":"+t.options["port"])
	if err != nil {
		fmt.Println(err) //signal not running
		return
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go func() {
			a := newAgent(conn, t.tcpResponseChan)
			t.scheduleCheckins(a, conn)
			t.agents[a.GetID()] = a
			t.CheckinChan <- server.Checkin{
				UID:          a.GetID(),
				ListenerName: t.GetName(),
				CheckinTime:  time.Now(),
			}
		}()
	}
}

func (t *JaqenTCPListener) listenCheckins() {
	tix := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-tix.C:
			for k, v := range t.agents {
				if !v.Running {
					delete(t.agents, k)
					continue
				}
				t.CheckinChan <- server.Checkin{
					UID:          v.GetID(),
					ListenerName: t.GetName(),
					CheckinTime:  time.Now(),
				}

			}
		}
	}
}

func (t *JaqenTCPListener) scheduleCheckins(a *agent, c net.Conn) {

}

func (t JaqenTCPListener) GetName() string {
	return "tcp"
}

func (t *JaqenTCPListener) SetOption(k, v string) {
	t.optionsMut.Lock()
	defer t.optionsMut.Unlock()
	t.options[k] = v
}

func (t *JaqenTCPListener) GetOption(s string) (string, error) {
	t.optionsMut.RLock()
	defer t.optionsMut.RUnlock()
	if x, ok := t.options[s]; ok {
		return x, nil
	}
	return "", errors.New("Option not found")
}

func (t *JaqenTCPListener) GetOptions() func(string) []string {
	t.optionsMut.RLock()
	defer t.optionsMut.RUnlock()
	return func(line string) []string {
		a := make([]string, 0)
		for k, _ := range t.options { //other options go here?
			a = append(a, k)
		}
		return a
	}
}

func (t *JaqenTCPListener) Kill(id string) {

}

func (t *JaqenTCPListener) SendCommand(agent, command string) {
	if string(command[len(command)-1]) != "\n" {
		command = command + "\n"
	}
	t.agents[agent].inChan <- command
}

func (t *JaqenTCPListener) Start() error {
	p, _ := t.GetOption("port")
	fmt.Println("Starting TCP Listerner on port " + p)
	go t.listen()
	t.running = true
	return nil
}

func (t *JaqenTCPListener) Stop() {}

func (t JaqenTCPListener) GenerateAgentFormats() []string { return []string{} }
func (t JaqenTCPListener) Generate(string) []byte         { return []byte{} }

func (t JaqenTCPListener) GetInfo() server.ListenerInfo {

	//copy the options map to avoid bad stuff
	t.optionsMut.Lock()
	newMap := make(map[string]string)
	for k, v := range t.options {
		newMap[k] = v
	}
	t.optionsMut.Unlock()

	return server.ListenerInfo{
		Name:    t.GetName(),
		Running: t.running,
		Options: newMap,
	}
}
