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

type tcpResponse struct {
	UID  string
	resp byte
}

type JaqenTCPListener struct {
	options         server.ListenerOptions
	tcpResponseChan chan tcpResponse
	CheckinChan     chan server.Checkin
	ResponseChan    chan server.Command

	agents          map[string]*agent
	aMut            *sync.RWMutex
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
	t.options = server.ListenerOptions{}.New()
	t.pendingCommands = make(map[string][]string)
	t.pcMut = &sync.RWMutex{}
	t.aMut = &sync.RWMutex{}
	t.CheckinChan = make(chan server.Checkin, 10)
	t.ResponseChan = make(chan server.Command, 10)
	t.tcpResponseChan = make(chan tcpResponse, 1)
	t.options.Set("port", "4444")
	t.options.Set("ip", "0.0.0.0")

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
	port, _ := t.options.Get("port")
	ip, _ := t.options.Get("ip")
	l, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		fmt.Println(err) //signal not running
		t.Stop()
		return
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go func() {
			a := newAgent(conn, t.tcpResponseChan)
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

func (t *JaqenTCPListener) listenCheckins() {
	tix := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-tix.C:
			t.aMut.Lock()
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
			t.aMut.Unlock()
		}
	}
}

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

func (t JaqenTCPListener) GetName() string {
	return "tcp"
}

func (t *JaqenTCPListener) SetOption(k, v string) {
	t.options.Set(k, v)
}

func (t *JaqenTCPListener) GetOption(s string) (string, error) {
	if x, ok := t.options.Get(s); ok {
		return x, nil
	}
	return "", errors.New("Option not found")
}

func (t *JaqenTCPListener) GetOptions() func(string) []string {
	return t.options.GetKeys()
}

func (t *JaqenTCPListener) Kill(id string) {

}

func (t *JaqenTCPListener) SendCommand(agent, command string) {
	//ensure commands begin with a newline
	if string(command[len(command)-1]) != "\n" {
		command = command + "\n"
	}
	t.aMut.Lock()
	t.agents[agent].inChan <- command
	t.aMut.Unlock()
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
	newMap := t.options.CloneMap()

	return server.ListenerInfo{
		Name:    t.GetName(),
		Running: t.running,
		Options: newMap,
	}
}
