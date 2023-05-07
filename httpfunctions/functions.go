package httpfunctions

import (
	//     "encoding/json"
	//     "fmt"
	//     "log"
	"bytes"
	"net/http"

	// "net/url"

	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
	// color "github.com/fatih/color"
	// gui "github.com/grupawp/warships-lightgui"
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

type Fire struct {
	Coord string `json:"coord"`
}

type Fire_result struct {
	Result string `json:"result"`
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

	time.Sleep(time.Second * 4)
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

	ui := gui.NewGUI(true)

	txt := gui.NewText(1, 1, "Press on any coordinate to log it.", nil)
	ui.Draw(txt)
	ui.Draw(gui.NewText(1, 2, "Press Ctrl+C to exit", nil))
	ui.Draw(gui.NewText(1, 3, data.Nick+" vs "+data.Opponent, nil))

	// boardConfig := gui.BoardConfig{RulerColor: gui.Color{Red: 236, Green: 54, Blue: 54}, TextColor: gui.Color{Red: 88, Green: 243, Blue: 212}}
	board := gui.NewBoard(1, 5, nil)
	board2 := gui.NewBoard(50, 5, nil)
	ui.Draw(board)
	ui.Draw(board2)
	states := [10][10]gui.State{}
	states2 := [10][10]gui.State{}
	for i := range states {
		states[i] = [10]gui.State{}
	}

	for _, coordinate := range bo.Board {
		x, y := ChangeCooerdinate(coordinate)
		states[x][y-1] = gui.Ship
	}
	board.SetStates(states)

	ui.Log("FIRE: %s", data)
	myHitCounter := 0
	oppHitCounter := 0

	if len(data.Opp_shots) != 0 {

		ui.Log("DATA: %s", data)
		x, y := ChangeCooerdinate(data.Opp_shots[len(data.Opp_shots)-1])
		if states[x][y-1] == gui.Ship {
			oppHitCounter++
			states[x][y-1] = gui.Hit
			time.Sleep(time.Second * 1)
			Get_Opp_Shot(board, board2, &states, &states2, ui, client, token, &oppHitCounter)
			if oppHitCounter == 20 {
				txt.SetText(fmt.Sprintf("OPPONENT WON: %s", data.Opponent))
			}
		} else {
			states[x][y-1] = gui.Miss
		}
		board.SetStates(states)
	}

	go func() {
		for {
			Shot(board, board2, &states, &states2, ui, client, token, &myHitCounter)
			if myHitCounter == 20 {
				txt.SetText(fmt.Sprintf("YOU WON: %s", data.Nick))
				break
			}
			Get_Opp_Shot(board, board2, &states, &states2, ui, client, token, &oppHitCounter)
			if oppHitCounter == 20 {
				txt.SetText(fmt.Sprintf("OPPONENT WON: %s", data.Opponent))
				break
			}
		}
	}()
	ui.Start(nil)
	// Get_Opp_Shot(board, board2, &states, &states2, ui, client, token)

	// Shot(board, board2, &states, &states2, ui, client, token)
	// go func() {
	// 	for {
	// 		char := board2.Listen(context.TODO())
	// 		txt.SetText(fmt.Sprintf("Coordinate: %s", char))
	// 		ui.Log("Coordinate: %s", char) // logs are displayed after the game exits
	// 		x, y := ChangeCooerdinate(char)
	// 		states2[x][y-1] = gui.Hit
	// 		board2.SetStates(states2)
	// 	}
	// }()

	// ui.Start(nil)

	// bb := gui.New(
	// 	gui.ConfigParams().
	// 		HitChar('#').
	// 		HitColor(color.FgRed).
	// 		BorderColor(color.BgRed).
	// 		RulerTextColor(color.BgYellow).
	// 		NewConfig())

	// bb.Import(bo.Board)
	// bb.Display()

	// fmt.Print(data.Nick + " vs " + data.Opponent)

	return nil
}

func ChangeCooerdinate(coordinate string) (int, int) {
	x_letter := coordinate[:1]
	y, err := strconv.Atoi(coordinate[1:])
	if err != nil {
		return -1, -1
	}
	x := 0
	switch x_letter {
	case "A":
		x = 0
	case "B":
		x = 1
	case "C":
		x = 2
	case "D":
		x = 3
	case "E":
		x = 4
	case "F":
		x = 5
	case "G":
		x = 6
	case "H":
		x = 7
	case "I":
		x = 8
	case "J":
		x = 9
	}
	return x, y
}

func Shot(myBoard, oppBoard *gui.Board, myStates, oppStates *[10][10]gui.State, ui *gui.GUI, client *http.Client, token string, myHitCounter *int) {
	var x, y int
	char := oppBoard.Listen(context.TODO())
	ui.Log("My Shot: %s", char)
	x, y = ChangeCooerdinate(char)
	fire_data := Fire{Coord: char}
	encoded_data, _ := json.Marshal(fire_data)

	req, err := http.NewRequest("POST", "https://go-pjatk-server.fly.dev/api/game/fire", bytes.NewBuffer(encoded_data))
	req.Header.Add("x-auth-token", token)
	if err != nil {
		ui.Log("cannot send data: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		ui.Log("cannot send data: %w", err)
	}

	var data Fire_result

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		ui.Log("cannot unmarshal data: %w", err)
	}
	ui.Log("result: %s", data.Result)
	if data.Result == "miss" {
		oppStates[x][y-1] = gui.Miss
		oppBoard.SetStates(*oppStates)
	} else {
		oppStates[x][y-1] = gui.Hit
		*myHitCounter++
		if *myHitCounter == 20 {
			return
		}
		oppBoard.SetStates(*oppStates)
		Shot(myBoard, oppBoard, myStates, oppStates, ui, client, token, myHitCounter)
	}
}

func Get_Opp_Shot(myBoard, oppBoard *gui.Board, myStates, oppStates *[10][10]gui.State, ui *gui.GUI, client *http.Client, token string, oppHitCounter *int) {
	var data GameStatus
	// for len(data.Opp_shots) != *oppHitCounter+1 {
	for len(data.Opp_shots) == 0 {
		time.Sleep(time.Millisecond * 200)
		req, err := http.NewRequest(http.MethodGet, "https://go-pjatk-server.fly.dev/api/game", nil)
		req.Header.Set("x-auth-token", token)
		resp, err2 := client.Do(req)
		if err != nil || err2 != nil {
			ui.Log("cannot create request: %w", err)
		}

		err3 := json.NewDecoder(resp.Body).Decode(&data)
		if err3 != nil {
			ui.Log("cannot unmarshal data: %w", err)
		}
	}

	ui.Log("DATA: %s", data)
	x, y := ChangeCooerdinate(data.Opp_shots[len(data.Opp_shots)-1])
	if myStates[x][y-1] == gui.Ship {
		myStates[x][y-1] = gui.Hit
		*oppHitCounter++
		if *oppHitCounter == 20 {
			return
		}
		time.Sleep(time.Second * 1)
		Get_Opp_Shot(myBoard, oppBoard, myStates, oppStates, ui, client, token, oppHitCounter)
	} else {
		myStates[x][y-1] = gui.Miss
	}
	myBoard.SetStates(*myStates)
}
