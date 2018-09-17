# Jaqen

"Speak the name, and death will come"

## Installation

Minimum Go version of 1.10 (may work on earlier versions too I guess)

```
go get -u github.com/c-sto/jaqen
```

Alternatively:

```
cd $GOPATH/src/github.com/
mkdir c-sto
cd c-sto
git clone https://github.com/c-sto/Jaqen
cd Jaqen
go get .
go run Jaqen.go
```

## Usage

Create a listener, set the listener settings, and generate an agent to deploy onto your already compromised host.

There are two listeners included - DNS and TCP. Additional listeners will be added in time, but they are intended to only be a template. Successful red teaming will require custom listeners/agents to achieve objectives. Basic AV evasion techniques are displayed in the DNS golang agent.

## Custom Listener

The listener plugs into the main C2 that you control via the CLI. The listener simply has to conform to the 'Listener' interface. The interface can be seen in the [interface](libJaqen/server/interface.go) source file. Any 'struct' type that implements every one of the functions defined in the interface will conform to the interface, and you will be able to add it to the [cli](cli/cli.go) file.

### Events

Events are passed back to the cli/server via channels - they are defined in the [interface](libJaqen/server/interface.go) file. The bare minimum required is to pass a uid back to the cli on checkin, and ideally some sort of response confirmation if a command has been executed, but the only limit is your creativity. Checkins _can_ have extra data (OS, agent type etc), but the only required field is the UID. Listeners handle their own agent UID's.

### Agents

Agents can do whatever you'd like. The DNS listener has bash, powershell, and golang agents provided as an example of how flexible it can be. The TCP listener can be used by simply sending a regular revese shell back (`metasploit shell_reverse_tcp`, `nc -e /bin/bash`, etc). Templating is encouraged to allow settings to be passed to agents. See the DNS listener for examples.


## Thanks

Inspiration for this was taken from merlin, http/2 c2 built by Ne0nD0g. Please go and look at it, it's very good.
https://github.com/Ne0nD0g/merlin