package cli

import (
	"strings"

	"github.com/chzyer/readline"
)

var nav map[string]readline.PrefixCompleter

func getCompleter(completer string) *readline.PrefixCompleter {
	s := readline.NewPrefixCompleter()
	s.GetName()
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
