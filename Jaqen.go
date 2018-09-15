package main

import (
	"math/rand"
	"time"

	//_ "net/http/pprof"

	"github.com/c-sto/Jaqen/cli"
	"github.com/fatih/color"
)

func main() {

	color.New().Println(`
....''.....'''''''''''''......'''''';o0KK00OOOkkkk
''''''''''''''''..'...'''''''''''''';oOKK00OOOOkkk
.'''''''''''''''........'''...'''..';oOKK0OOOOOOkk
..........''''.....'''''.......'''..,oOKK0OOOOOOOO
............... 'lxkkkkxolc,...,,,'..:k000OOOOOOOO
.............. .lxkkOOkkdddc'...'','..lO00OOOOOOOO
.............  'loddxxxxdddo;. ...';'..oOOOkOOOkkO
...''''......  'lodxxdddoool;.  ...,;..'dOOOOOkkkk
''''''''''...  .;;;col:,;;;,..  .''.,'..,dOOOOkkkk
''''''''''..   .''.cxo:,;:;'..  .';;;;'..;xOOkkxxx
''''''''''..   .;:lkkxolllooo:. ..,:::;...lkOOkxdx
''''''''''..   .:lodxdodxxxdoc....;cl:,...'dOkkxdd
'''..'''''..    'cc;;:lodddol;...':ld:.....ckkkxdd
'''''...'...    .;;,,;::clll;'..';:od:'...'lkOOkxd
''''''''....     .;;,;ccllc;,...,,,:ll::'.,lxkOkxd
'''''''''...      .;:clllc:,....',,'':ll:,,:lxkkxd
''''',,''''...     .',;,'.........'..,ll;,:ccdkkxx
,,',,,,,'''....      .''..'',,..  .'..',,,:odxkxxd
'''''',,,,'...       .,:::c:;,'. .....,cxkOOOkxxdo
.'''''''''''....    ..';::::,,,.. .':okO00KKK0Okxo
'',,,,''........  ..',,,;;;::;;:;;cdkO000000KKK0ko
',;;;;,,''........'';;;;::cc::cllodxkOO0KKKKKKKKKk
,,;;;;;,,'''''',;;;::;::ccc:clclooxkkkkOO00KKXXKKK
',,,,,;,,'',,,,cocc::cccc::lolooooc:;,;;::cok0KKKK
',,,',,,,''',,:odoccccclccldodoc;...'';;,''';lOKKK

`)
	color.New().Println(randomQuote())

	cli.Shell()

}

func randomQuote() string {
	rand.Seed(time.Now().UnixNano())
	q := []string{
		"Valar morghulis",
	}
	return q[rand.Intn(len(q))]
}
