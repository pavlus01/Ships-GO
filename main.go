package main

import (
	"log"
	"shipsgo/httpfunctions"
)

func main() {
	err := httpfunctions.FirstConnection("Kapitan to ja!", "kapitan")
	if err != nil {
		log.Fatal(err)
	}
}
