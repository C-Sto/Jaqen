package tcpListener

import (
	"bufio"
	"fmt"
	"io"
	"net"

	uuid "github.com/satori/go.uuid"
)

type agent struct {
	inChan     chan string      //data to send
	outChan    chan tcpResponse //data received
	reader     *bufio.Reader
	writer     *bufio.Writer
	conn       net.Conn
	connection *agent
	id         string
	Running    bool
}

func (a *agent) GetID() string {

	if a.id == "" {
		//generate id
		x, err := uuid.NewV4()
		if err != nil {
			fmt.Println(err)
		}
		a.id = x.String()
		//a.id = "cats"
	}
	return a.id
}

func (a *agent) Read() {
	smallbuf := make([]byte, 1)
	for {
		//read 1 byte at a time because reasons
		_, err := a.reader.Read(smallbuf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Agent disconnected")
				a.Running = false
				break
			}
			fmt.Println("Agent read error:")
			fmt.Println(err)
			break
		}
		a.outChan <- tcpResponse{UID: a.GetID(), resp: smallbuf[0]}

	}
	a.conn.Close()
	if a.connection != nil {
		a.connection.connection = nil
	}
	a = nil
}

func (a *agent) Write() {
	for data := range a.inChan {
		a.writer.WriteString(data)
		a.writer.Flush()
	}
}

func (a *agent) Listen() {
	a.Running = true
	go a.Read()
	go a.Write()
}
