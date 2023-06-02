package main

import (
	"log"
	"net/http"
	"shipsgo/game"
	"shipsgo/intercation"
	"time"
)

func main() {

	client := &http.Client{Timeout: time.Second * 5}

	nick, desc := intercation.PlayerDescription()
	var nick_pointer *string = &nick

	for {
		coord := intercation.OwnBoard()
		oppo, err := intercation.ShowPlayersList(*client)

		if err != nil {
			log.Fatal(err)
		}

		err = game.Game(desc, nick_pointer, oppo, *client, coord)

		if err != nil {
			log.Fatal(err)
		}

		err = intercation.PostGameStatistics(nick_pointer, *client)

		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(5 * time.Second)
	}
}
