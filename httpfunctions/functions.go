package httpfunctions

import (
	//     "encoding/json"
	//     "fmt"
	//     "log"
	"bytes"
	"net/http"

	// "net/url"

	// "context"
	"encoding/json"
	"fmt"
	"time"

	// gui "github.com/grupawp/warships-gui/v2"
	color "github.com/fatih/color"
	gui "github.com/grupawp/warships-lightgui"
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
	Timer            int      `json:"timer"`
}

type GameStatus struct {
	Nick             string   `json:"nick"`
	Game_status      string   `json:"game_status"`
	Last_game_status string   `json:"last_game_status"`
	Opponent         string   `json:"opponent"`
	Opp_shots        []string `json:"opp_shots"`
	Should_fire      bool     `json:"should_fire"`
	Timer            int      `json:"timer"`
}

type Board struct {
	Board []string `json:"board"`
}

func FirstConnection(desc, name string) error {
	request_data := Request_data{Desc: desc, Nick: name, Target_nick: "", Wpbot: true}
	encoded_data, _ := json.Marshal(request_data)
	// log.Print(string(encoded_data))

	resp, err := http.Post("https://go-pjatk-server.fly.dev/api/game", "application/json", bytes.NewBuffer(encoded_data))
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}
	token := resp.Header.Get("x-auth-token")
	// log.Print("Token " + token)

	time.Sleep(time.Second * 3)
	client := &http.Client{Timeout: time.Second * 5}
	req, err2 := http.NewRequest(http.MethodGet, "https://go-pjatk-server.fly.dev/api/game", nil)
	req.Header.Set("x-auth-token", token)
	resp2, err3 := client.Do(req)
	if err2 != nil || err3 != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	var data GameStatus
	// resBody, err := ioutil.ReadAll(resp2.Body)
	// if err != nil {
	// 	fmt.Printf("client: could not read response body: %s\n", err)
	// 	os.Exit(1)
	// }
	// log.Print(string(resBody))
	err = json.NewDecoder(resp2.Body).Decode(&data)
	if err != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}
	// log.Print("Game Status " + data.Game_status)
	// log.Print("YOUR NICK " + data.Nick)
	// log.Print(data)

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

	// ui := gui.NewGUI(true)

	// txt := gui.NewText(1, 1, "Press on any coordinate to log it.", nil)
	// ui.Draw(txt)
	// ui.Draw(gui.NewText(1, 2, "Press Ctrl+C to exit", nil))
	// ui.Draw(gui.NewText(1, 3, data.Nick+" vs "+data.Opponent, nil))

	// // boardConfig := gui.BoardConfig{RulerColor: gui.Color{Red: 236, Green: 54, Blue: 54}, TextColor: gui.Color{Red: 88, Green: 243, Blue: 212}}
	// board := gui.NewBoard(1, 5, nil)
	// board2 := gui.NewBoard(5, 10, nil)
	// ui.Draw(board)
	// ui.Draw(board2)

	// go func() {
	// 	for {
	// 		char := board.Listen(context.TODO())
	// 		txt.SetText(fmt.Sprintf("Coordinate: %s", char))
	// 		ui.Log("Coordinate: %s", char) // logs are displayed after the game exits
	// 	}
	// }()

	// gui.Start(nil)

	fmt.Print(data.Nick + " vs " + data.Opponent + "\n\n\n")

	bb := gui.New(
		gui.ConfigParams().
			HitChar('#').
			HitColor(color.FgRed).
			BorderColor(color.BgRed).
			RulerTextColor(color.BgYellow).
			NewConfig())

	bb.Import(bo.Board)
	bb.Display()

	return nil
}
