package cli

import (
	"io"
	"strings"
	"sync"

	"github.com/c-sto/Jaqen/libJaqen/server"
	"github.com/chzyer/readline"
)

type cliState struct {
	context      string
	ContextMutex *sync.RWMutex
	prompt       *readline.Instance
	localServer  *server.JaqenServer

	SelectedAgent string

	outReader            chan server.Output
	selectedListener     string //actual name of listener ("dns:asdf" etc)
	selectedListenerType string //type of listener ("dns" etc)
	tempListener         server.Listener

	agentToAlias map[string]string
	aliasToAgent map[string]string
	aliasMutex   *sync.RWMutex
}

func (s cliState) GetContext() string {
	s.ContextMutex.RLock()
	defer s.ContextMutex.RUnlock()
	return s.context
}

func (s *cliState) SetContext(c, prompt string) {
	s.ContextMutex.Lock()
	defer s.ContextMutex.Unlock()
	s.context = c
	s.prompt.SetPrompt(prompt + "> ")
	s.prompt.Config.AutoComplete = getCompleter(strings.ToLower(c))
}

func (s *cliState) SetServer(srv *server.JaqenServer) {
	s.ContextMutex.Lock()
	defer s.ContextMutex.Unlock()
	s.localServer = srv
}

func (s cliState) New() cliState {
	c := cliState{
		//Context:      "main",
		ContextMutex: &sync.RWMutex{},
		//prompt * readline.Instance,
	}
	ls := server.JaqenServer{}
	ls.Init()
	c.SetServer(&ls)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "> ",
		HistoryFile:       "logs.txt",
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		AutoComplete:      getCompleter("main"),
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}

	c.aliasToAgent = make(map[string]string)
	c.agentToAlias = make(map[string]string)
	c.aliasMutex = &sync.RWMutex{}

	c.SetPromptFullConfig(rl)
	c.SetContext("main", "")

	c.outReader = c.localServer.GetOutputChan()
	go c.StartOutReader()

	return c
}

func (s *cliState) AgentExecute(sa, cmd string) {
	if a, ok := s.aliasToAgent[sa]; ok {
		s.localServer.AgentExecute(a, cmd)
	} else {
		s.localServer.AgentExecute(sa, cmd)
	}
}

func (s *cliState) SetAlias(agent, alias string) error {
	s.aliasMutex.Lock()
	defer s.aliasMutex.Unlock()

	//todo: confirm they exist
	//todo: confirm no overwrite
	s.agentToAlias[agent] = alias
	s.aliasToAgent[alias] = agent

	return nil
}

func (s cliState) ListListeners(string) []string {
	r := []string{}

	for _, x := range s.localServer.GetListeners() {
		r = append(r, x.Name)
	}
	return r
}

func (s *cliState) StartOutReader() {
	//green := color.New(color.FgGreen).SprintFunc()
	for {
		select {
		case outLine := <-s.outReader:
			io.WriteString(s.prompt.Stdout(), outLine.Val+"\n")
		}
	}
}

func (s *cliState) SetPromptFullConfig(i *readline.Instance) {
	s.prompt = i
}

func (s *cliState) CloseReadline() {
	s.prompt.Close()
}

func (s *cliState) ReadLine() (string, error) {
	return s.prompt.Readline()
}
