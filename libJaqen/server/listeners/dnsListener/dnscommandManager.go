package dnsListener

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"sync"
)

/*
type dnsresponse struct {
	Chunks      []dnschunk
	TotalChunks int64
	ReadChunks  int64
}*/

type dnsresponse struct {
	UID         string
	MaxChunks   string
	ThisChunk   string
	Payload     string
	CmdID       string
	Response    []byte
	Chunks      []dnschunk
	TotalChunks int64
	ReadChunks  int64
}

type dnschunk struct {
	Body string
	Num  int64
}

type DNSCommand struct {
	Command  string
	UID      string
	CmdID    string
	Response dnsresponse
}

type dnscommandManager struct {
	commandMap     map[string]DNSCommand
	cMMutex        *sync.RWMutex
	waitingCommand string
}

func (cm *dnscommandManager) Init() {
	cm.commandMap = make(map[string]DNSCommand)
	cm.cMMutex = &sync.RWMutex{}

}

func (cm dnscommandManager) GetCommandToSend() string {
	cm.cMMutex.RLock()
	defer cm.cMMutex.RUnlock()
	return cm.waitingCommand
}

func (cm *dnscommandManager) GetCommand(c string) DNSCommand {
	cm.cMMutex.RLock()
	defer cm.cMMutex.RUnlock()
	if v, ok := cm.commandMap[c]; ok {
		return v
	}
	return DNSCommand{}
}

func (cm *dnscommandManager) ClearCommand(c string) {
	cm.cMMutex.Lock()
	defer cm.cMMutex.Unlock()
	delete(cm.commandMap, c)
}

func (cm *dnscommandManager) SetCommandToSend(s string) {
	cm.cMMutex.Lock()
	defer cm.cMMutex.Unlock()
	cm.waitingCommand = s
}

//uuid == cmdid
func (cm *dnscommandManager) UpdateCmd(uuid, maxchunks, thischunk, vals string) {
	cm.cMMutex.Lock()
	defer cm.cMMutex.Unlock()
	c, ok := cm.commandMap[uuid]

	if !ok {
		return
	}

	cn, e := strconv.ParseInt(thischunk, 10, 64)
	if e != nil {
		fmt.Println("Bad chunk number: ", e)
		return
	}
	mc, e := strconv.ParseInt(maxchunks, 10, 64)
	if e != nil {
		fmt.Println("Bad max chunk number: ", e)
		return
	}
	c.Response.TotalChunks = mc
	c.Response.AddChunk(cn, vals)
	c.Response.ReadChunks++
	cm.commandMap[uuid] = c

	fmt.Println(fmt.Sprintf("Recv %d of %d", c.Response.ReadChunks, c.Response.TotalChunks))

}

func (cm *dnscommandManager) AddCommand(c DNSCommand) {
	cm.cMMutex.Lock()
	defer cm.cMMutex.Unlock()
	cm.commandMap[c.UID] = c
}

func (r *dnsresponse) AddChunk(cnum int64, val string) {
	r.Chunks = append(r.Chunks, dnschunk{Body: val, Num: cnum})
}

func (cm *dnscommandManager) IsDone(cmdId string) bool {
	c := cm.GetCommand(cmdId).Response
	return c.IsDone()
}

func (r dnsresponse) IsDone() bool {
	if r.TotalChunks == 0 || r.TotalChunks > r.ReadChunks {
		return false
	}
	return true
}

func (r dnsresponse) ReadResposne() string {
	//sort chunks
	rval := ""
	sort.Slice(r.Chunks, func(i, j int) bool {
		if r.Chunks[i].Num < r.Chunks[j].Num {
			return true
		}
		return false
	})
	for _, x := range r.Chunks {
		rval += x.Body
	}

	v, e := hex.DecodeString(rval)
	if e != nil {
		fmt.Println("ReadResponse Error:", e)
	}
	return string(v)
}
