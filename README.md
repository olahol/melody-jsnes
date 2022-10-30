# melody-jsnes

> Super simplistic Multiplayer NES server in Go.

melody-jsnes is a demo project showing off Go's real-time web app
capabilities. Its design is straight forward, it just snapshots the
canvas of player one and sends it to player two and sends back inputs
from player two. Images data goes in direction, key codes in the other.

![demo](https://cdn.rawgit.com/olahol/melody-jsnes/master/demo.gif "Me playing a perfectly legal version of Contra with my friends")

## Usage

You will need to have at least one NES ROM with the extension `.nes`.

    $ git clone --recursive https://github.com/olahol/melody-jsnes
    $ go get
    $ go build
    $ ./melody-jsnes game.nes
    $ $BROWSER http://localhost:5000

## Contributors

* Ola Holmström (@olahol)
* Chris Cacciatore (@cacciatc)
