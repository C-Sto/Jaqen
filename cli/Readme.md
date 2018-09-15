# Server CLI

Basic structure (intend to update this, but it will probably get left l0l) (x indicates done/working)

```
Main
|-Exit (Exits the Server, shuts down agents etc) X
|
|-Listeners (Listener management) X
|   |-Create (Add new listener) X
|   |-Interact <listener> (Edit selected listener)
|   |       |-Add (Add listener to list)
|   |       |-Set (Set options)
|   |       |-Show (Show options)
|   |       |-Start (Start and add to list)
|   |       |-Back (don't add to list)
|   |-Delete (Delete current listener)
|   |-Show (Show current listeners)
|   |-Stop (Stop listener)
|
|-Agents (Agent Management) X
|   |-Show (Show current agents) X
|   |-Generate (Create an agent(undecided if this should be a feature))
|   |   |-Set (Set options)
|   |   |-Show [options/info]
|   |   |-Generate
|   |-Set (set agent options, output location etc)
|   |-Interact (Interact with an agent) X
|   |   |-Exec (command execution) X
|   |   |   |-Send command (Send a command to an agent)
|   |   |   |-Send shellcode (Send shellcode to an agent)
|   |   |-Show output (Show log of agent historic output)
|   |   |-Kill (Kill the agent)
|   |-Back
|
|-Version (Version info) X
|-Status (Show current agents, current listeners)
```