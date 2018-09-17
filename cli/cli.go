package cli

import (
	"fmt"
	"os"
	"strings"

	//. "github.com/c-sto/Jaqen"
	"github.com/c-sto/Jaqen/libJaqen/server"
	"github.com/c-sto/Jaqen/libJaqen/server/listeners/dnsListener"
	"github.com/c-sto/Jaqen/libJaqen/server/listeners/encryptedDNSListener"
	"github.com/c-sto/Jaqen/libJaqen/server/listeners/tcpListener"
	"github.com/c-sto/readline"
	"github.com/fatih/color"
)

var state cliState

var definedListeners = map[string]server.Listener{
	//add listener here once written
	"dns":        &dnsListener.JaqenDNSListener{},
	"secure-dns": &encryptedDNSListener.JaqenEncryptedDNSListener{},
	"tcp":        &tcpListener.JaqenTCPListener{},
}

func DoTest() {
	/*
		l := dnsListener.JaqenDNSListener{}
		state.localServer.AddListener(&l)
		//fmt.Println(s.GetOptions("dns")("""))
		state.localServer.SetOption("dns", "domain", "c2.supershady.ru")
		fmt.Println(state.localServer.GetOption("dns", "domain"))
		fmt.Println(state.localServer.Start("dns"))
		state.localServer.GenerateAgent("dns", "powershell")
		state.selectedListener = "dns"
		state.SetContext("listenerinteract", "Listeners>Interact[dns]")
	*/
	/*
		t := tcpListener.JaqenTCPListener{}
		state.localServer.AddListener(&t)
		state.localServer.SetOption("tcp", "port", "4444")
		state.localServer.Start("tcp")
		//*/
}

func printVersion() string {

	return "0.0.1"
}

func printStatus() string {
	return "NO"
}

// Shell is the exported function to start the command line interface
func Shell() {

	state = cliState{}.New()
	state.SetContext("main", "")

	DoTest()

	defer state.CloseReadline()

	for {
		line, err := state.ReadLine() // rl.Readline()
		if err != nil {
			break
		}
		if len(line) < 1 {
			continue
		}
		line = strings.TrimSpace(line)

		cmd := strings.Fields(line)
		keyVal := strings.ToLower(cmd[0])
		switch strings.ToLower(state.GetContext()) {
		case "main":
			switch keyVal {
			case "exit":
				exit()
			case "listeners":
				state.SetContext("Listeners", "Listeners")
			case "agents":
				state.SetContext("Agents", "Agents")
			case "version":
				printVersion()
			case "status":
				printStatus()
			}
		case "listeners":
			switch keyVal {
			case "create":
				state.SetContext("listenercreate", "Listeners>Create")
			case "interact":
				//require more than 0 args
				if len(cmd) != 2 {
					fmt.Println("Interact <listener>")
					continue
				}
				state.selectedListener = cmd[1]
				state.SetContext("listenerinteract", fmt.Sprintf("Listeners>Interact[%s]", yellow(state.selectedListener)))
			case "delete":
				if len(cmd) > 1 {
					state.localServer.RemoveListener(cmd[1])
				}
			case "show":
				//get all listeners
				x := state.localServer.GetListeners()
				for _, y := range x {
					opts := ""
					for k, v := range y.Options {
						opts += fmt.Sprintf(" %s=%s", k, v)
					}
					fmt.Printf("Name: %s Running: %v Options: %s\n", y.Name, y.Running, opts)
				}
			case "stop":
				if len(cmd) > 1 {
					//todo: check listener exists
					state.localServer.StopListener(cmd[1])
				}
			case "back":
				state.SetContext("main", "")
			}
		case "listenercreate": //select listener to create (select listener type)
			switch keyVal {
			default:
				if _, ok := definedListeners[keyVal]; ok {
					state.selectedListenerType = keyVal
					state.tempListener = definedListeners[keyVal]

					n, e := state.localServer.AddListener(state.tempListener)
					if e != nil {
						fmt.Println(e)
					}
					state.tempListener = nil
					state.selectedListener = n
					state.SetContext("listenerinteract", fmt.Sprintf("Listeners>Interact[%s]", yellow(state.selectedListener)))
				}
			case "back":
				state.SetContext("Listeners", "Listeners")
			}

		case "listenerinteract":
			switch keyVal {
			case "start":
				e := state.localServer.Start(state.selectedListener)
				if e != nil {
					fmt.Println(e)
				}
			case "info":
				x, _ := state.localServer.GetOptions(state.selectedListener)
				fmt.Println(x("")) //wow this is hella gross.
			case "set":
				if len(cmd) == 3 {
					state.localServer.SetOption(state.selectedListener, strings.ToLower(cmd[1]), cmd[2])
				} else {
					fmt.Println("Require 3 arguments: Set <option> <value>")
					continue
				}
			case "generate":
				if len(cmd) > 1 {
					fmt.Println(string(state.localServer.GenerateAgent(state.selectedListener, strings.ToLower(cmd[1]))))
				} else {

				}
			case "back":
				state.tempListener = nil
				state.selectedListener = ""
				state.SetContext("listeners", "Listeners")
			}

		case "agents":
			switch keyVal {
			case "show":
				fmt.Println(getAgents(""))
			case "set":
				//todo: this (enable writing to local file for agent output)
			case "interact":
				if len(cmd) < 2 {
					fmt.Println("Interact <agent>")
					continue
				}
				al := getAgents("")
				for x := range al {
					if al[x] == cmd[1] {
						state.SelectedAgent = cmd[1]
						state.SetContext("agentinteract", fmt.Sprintf("Agents>Interact[%s]", yellow(state.SelectedAgent)))
						continue
					}
				}
				fmt.Println("Agent not found")
			case "back":
				state.SetContext("main", "")
			}
		case "agentset":
			//todo: this
		case "agentinteract":
			switch keyVal {
			case "set":
				if len(cmd) > 1 {
					switch strings.ToLower(cmd[1]) {
					case "alias":
						//todo: check agent exists gracefully
						//todo: show error output for bad cmds
						if len(cmd) > 3 {
							state.SetAlias(cmd[2], cmd[3])
							state.SelectedAgent = cmd[3]
							state.SetContext("agentinteract", fmt.Sprintf("Agents>Interact[%s]", yellow(state.SelectedAgent)))
						}
					}
				}
			case "exec":
				if len(cmd) == 1 {
					state.SetContext("interactexec", fmt.Sprintf("Agents>Interact[%s]>%s", yellow(state.SelectedAgent), red("Exec")))
				} else {
					state.AgentExecute(state.SelectedAgent, strings.Join(cmd[1:], " "))
				}
			case "kill":
				if len(cmd) > 1 {
					state.localServer.AgentKill(cmd[1])
				}
			case "back":
				state.SetContext("Agents", "Agents")
			}
		case "interactexec":
			switch keyVal {
			default:
				state.AgentExecute(state.SelectedAgent, strings.Join(cmd[:], " "))
			case "back":
				state.SetContext("agentinteract", fmt.Sprintf("Agents>Interact[%s]", yellow(state.SelectedAgent)))
			}

		default:
			fmt.Println(keyVal)
			fmt.Println(strings.ToLower(state.GetContext()))
			//help?
		}
	}
}

var red = color.New(color.FgRed).SprintfFunc()
var yellow = color.New(color.FgYellow).SprintfFunc()

func getCompleter(completer string) *readline.PrefixCompleter {

	switch strings.ToLower(completer) {
	case "main":
		return readline.NewPrefixCompleter(
			readline.PcItem("Listeners"),
			readline.PcItem("Agents"),
			readline.PcItem("Status"),
			readline.PcItem("Version"),
			readline.PcItem("Exit"),
		)
	case "listeners":
		return readline.NewPrefixCompleter(
			readline.PcItem("Create"),
			readline.PcItem("Show"),
			readline.PcItem("Interact",
				readline.PcItemDynamic(state.ListListeners),
			),
			readline.PcItem("Stop",
				readline.PcItemDynamic(state.ListListeners),
			),
			readline.PcItem("Delete"),
			readline.PcItem("Back"),
		)
	case "agents":
		return readline.NewPrefixCompleter(
			readline.PcItem("Show"),
			readline.PcItem("Interact",
				readline.PcItemDynamic(getAgents),
			),
			readline.PcItem("Kill",
				readline.PcItemDynamic(getAgents),
			),
			readline.PcItem("Back"),
		)
	case "agentinteract":
		return readline.NewPrefixCompleter(
			readline.PcItem("Exec"),
			readline.PcItem("Show"),
			readline.PcItem("Set",
				readline.PcItem("Alias", readline.PcItemDynamic(getAgents)),
				readline.PcItem("OutLocation", readline.PcItem("Stdout"), readline.PcItem("File")),
			),
			readline.PcItem("Kill"),
			readline.PcItem("Back"),
		)
	case "listenercreate":
		return readline.NewPrefixCompleter(
			readline.PcItemDynamic(getDefinedListeners),
			readline.PcItem("Back"),
		)
	case "listenerinteract":
		x, e := state.localServer.GetOptions(state.selectedListener)
		if e != nil {
			x = func(string) []string {
				return []string{}
			}
		}
		return readline.NewPrefixCompleter(
			readline.PcItem("Start"),
			readline.PcItem("Info"),
			readline.PcItem("Set",
				readline.PcItemDynamic(x),
			),
			readline.PcItem("Generate",
				readline.PcItemDynamic(getAgentGenerateFormats(state.selectedListener)),
			),
			readline.PcItem("Restart"),
			readline.PcItem("Back"),
		)

	}
	return readline.NewPrefixCompleter(
		readline.PcItem("Listeners"),
		readline.PcItem("Agents"),
		readline.PcItem("Status"),
		readline.PcItem("Version"),
		readline.PcItem("Exit"),
	)
}

func getAgents(string) []string {
	a := state.localServer.GetAgents()("")
	//don't show any aliased agents
	o := []string{}
	for _, x := range a {
		if v, ok := state.agentToAlias[x]; ok {
			o = append(o, v)
		} else {
			o = append(o, x)
		}
	}
	return o
}

func getAgentGenerateFormats(l string) func(string) []string {

	r := func(string) []string {
		re := []string{}
		for _, x := range state.localServer.GetAgentFormats(l) {
			re = append(re, x)
		}
		return re
	}
	return r
}

func getDefinedListeners(string) []string {
	r := []string{}
	for k := range definedListeners {
		r = append(r, k)
	}
	return r
}

func exit() {
	os.Exit(0)
}
