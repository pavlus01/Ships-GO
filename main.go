package main

import (
	"log"
	"net/http"
	"shipsgo/httpfunctions"
	"shipsgo/intercation"
	"time"
)

func main() {

	client := &http.Client{Timeout: time.Second * 5}

	nick, desc := intercation.PlayerDescription()
	var nick_pointer *string = &nick
	for {
		oppo, err := intercation.ShowPlayersList(*client)
		coord := intercation.OwnBoard()
		err2 := httpfunctions.FirstConnection(desc, nick_pointer, oppo, *client, coord)
		err3 := intercation.PostGameStatistics(nick_pointer, *client)
		time.Sleep(5 * time.Second)

		if err != nil {
			log.Fatal(err)
		}
		if err2 != nil {
			log.Fatal(err2)
		}
		if err3 != nil {
			log.Fatal(err2)
		}
	}
}
