package httpfunctions

import (
	//     "encoding/json"
	//     "fmt"
	//     "log"
	"net/http"
	// "net/url"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Request_data struct {
	Desc        string `json:"desc"`
	Nick        string `json:"nick"`
	Target_nick string `json:"target_nick"`
	Wpbot       bool   `json:"wpbot"`
}

type StatusResponse struct {
	Desc             string   `json:"desc"`
	Game_status      string   `json:"game_status"`
	Last_game_status string   `json:"last_game_status"`
	Nick             string   `json:"nick"`
	Opp_desc         string   `json:"opp_desc"`
	Opp_shots        []string `json:"opp_shots"`
	Opponent         string   `json:"opponent"`
	Should_fire      bool     `json:"should_fire"`
	Timer            string   `json:"timer"`
}

type Board struct {
	Board []string `json:"board"`
}

func FirstConnection(desc, name string) error {
	request_data := Request_data{desc, name, "", true}
	encoded_data, _ := json.Marshal(request_data)

	resp, err := http.Post("https://go-pjatk-server.fly.dev/api/game", "application/json", bytes.NewBuffer(encoded_data))
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}
	token := resp.Header.Get("x-auth-token")
	// log.Print("Token " + token)

	req, err2 := http.NewRequest("GET", "https://go-pjatk-server.fly.dev/api/game", nil)
	req.Header.Add("x-auth-token", token)
	client := &http.Client{Timeout: time.Second * 5}
	resp2, err3 := client.Do(req)
	if err2 != nil || err3 != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	var data StatusResponse
	err = json.NewDecoder(resp2.Body).Decode(&data)
	if err != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}
	log.Print("Game Status " + data.Game_status)
	log.Print("YOUR NICK " + data.Nick)
	log.Print(data)

	req2, err4 := http.NewRequest("GET", "https://go-pjatk-server.fly.dev/api/game/board", nil)
	req2.Header.Add("x-auth-token", token)
	if err4 != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}
	resp4, err5 := client.Do(req2)
	if err5 != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}
	var bo Board
	err = json.NewDecoder(resp4.Body).Decode(&bo)
	if err != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}

	// log.Print(bo)

	// bb := gui.New(
	// 	gui.ConfigParams().
	// 		HitChar('#').
	// 		HitColor(color.FgRed).
	// 		BorderColor(color.BgRed).
	// 		RulerTextColor(color.BgYellow).
	// 		NewConfig())

	// bb.Import(bo.Board)
	// bb.Display()

	return nil
}
