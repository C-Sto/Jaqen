package encryptedDNSListener

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/c-sto/Jaqen/libJaqen/server/util"
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
	Key         []byte
}

type dnschunk struct {
	Body string
	Num  int64
}

type dnsCommand struct {
	Command  string
	UID      string
	CmdID    string
	Response dnsresponse
}

type dnsCommandManager struct {
	commandMap     map[string]dnsCommand
	cMMutex        *sync.RWMutex
	waitingCommand string
}

func (cm *dnsCommandManager) Init() {
	cm.commandMap = make(map[string]dnsCommand)
	cm.cMMutex = &sync.RWMutex{}

}

func (cm dnsCommandManager) GetCommandToSend() string {
	cm.cMMutex.RLock()
	defer cm.cMMutex.RUnlock()
	return cm.waitingCommand
}

func (cm *dnsCommandManager) GetCommand(c string) dnsCommand {
	cm.cMMutex.RLock()
	defer cm.cMMutex.RUnlock()
	if v, ok := cm.commandMap[c]; ok {
		return v
	}
	return dnsCommand{}
}

func (cm *dnsCommandManager) ClearCommand(c string) {
	cm.cMMutex.Lock()
	defer cm.cMMutex.Unlock()
	delete(cm.commandMap, c)
}

func (cm *dnsCommandManager) SetCommandToSend(s string) {
	cm.cMMutex.Lock()
	defer cm.cMMutex.Unlock()
	cm.waitingCommand = s
}

//uuid == cmdid
func (cm *dnsCommandManager) UpdateCmd(uuid, maxchunks, thischunk, vals string) {
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
	cm.commandMap[uuid] = c

	fmt.Println(fmt.Sprintf("Recv %d of %d", c.Response.ReadChunks, c.Response.TotalChunks))

}

func (cm *dnsCommandManager) AddCommand(c dnsCommand) {
	cm.cMMutex.Lock()
	defer cm.cMMutex.Unlock()
	cm.commandMap[c.UID] = c
}

func (r *dnsresponse) AddChunk(cnum int64, val string) {
	//check for existing chunk
	for _, x := range r.Chunks {
		if x.Num == cnum {
			return
		}
	}
	r.Chunks = append(r.Chunks, dnschunk{Body: val, Num: cnum})
	r.ReadChunks++
}

func (cm *dnsCommandManager) IsDone(cmdId string) bool {
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

	v, e := util.DecryptHexStringToString(rval, r.Key) //hex.DecodeString(rval)

	//util.Decrypt
	if e != nil {
		fmt.Println("ReadResponse Error:", e)
	}
	return string(v)
}
