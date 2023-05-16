package main

import (
	"log"
	"shipsgo/httpfunctions"
	"shipsgo/intercation"
)

func main() {

	nick, desc := intercation.PlayerDescription()
	oppo, err := intercation.ShowPlayersList()
	err2 := httpfunctions.FirstConnection(desc, nick, oppo)
	if err != nil {
		log.Fatal(err)
	}
	if err2 != nil {
		log.Fatal(err2)
	}
}
