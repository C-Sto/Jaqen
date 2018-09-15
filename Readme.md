# Jaqen

"Speak the name, and death will come"

Long term goal is a multilingual C2 built in golang. Agents simply need to conform to the communications spec, and can be written in any langauge.

For now, I just want a good DNS C2. 

Inspiration for this was taken from merlin, http/2 c2 built by Ne0nD0g. Please go and look at it, it's very good.
https://github.com/Ne0nD0g/merlin

Agents must be able to send:
- Agent checkin
- Agent heartbeat
- Command response (optional I guess)

Ideally, agents should be able to use multiple c2 domains (or fallbacks) in order to prevent getting cut off when the blue team notice something strange going on at superlegitdomain.ru.

Agents don't *Need* to be able to take a command and execute it. It's much more useful, but if you need to be stealthy, a daily heartbeat, and maybe some status update is probably fine (think a daily mimikatz to snarf up all the passwords logged that day).

If you build a jaqen agent, *please* make sure you have a killswitch. 

A jaqen agent is provided in golang. It will be the most feature rich since it will be developed alongside the server.

The goal is that agents should be flexible, provided they conform to the communication spec.