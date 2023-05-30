package httpfunctions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"shipsgo/helpers"
	"strconv"
	"strings"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

type Request_data struct {
	Coords      []string `json:"coords"`
	Desc        string   `json:"desc"`
	Nick        string   `json:"nick"`
	Target_nick string   `json:"target_nick"`
	Wpbot       bool     `json:"wpbot"`
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

type GameDescription struct {
	Desc     string `json:"desc"`
	Nick     string `json:"nick"`
	Opp_desc string `json:"opp_desc"`
	Opponent string `json:"opponent"`
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

func Game(desc string, name *string, opponent string, client http.Client, coord []string) error {
	var request_data Request_data
	if *name != "" {
		request_data = Request_data{Coords: coord, Desc: desc, Nick: *name, Target_nick: "", Wpbot: false}
	} else {
		request_data = Request_data{Coords: coord, Target_nick: "", Wpbot: false}
	}
	if opponent != "" {
		if opponent == "wpbot" {
			request_data.Target_nick = ""
			request_data.Wpbot = true
		} else {
			request_data.Target_nick = opponent
		}
	}
	encoded_data, _ := json.Marshal(request_data)

	resp, err := helpers.Request(client, http.MethodPost, "https://go-pjatk-server.fly.dev/api/game", bytes.NewBuffer(encoded_data), "", 5)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}
	token := resp.Header.Get("x-auth-token")

	var data GameStatus

	for data.Game_status != "game_in_progress" {

		time.Sleep(time.Second * 1)
		resp2, err2 := helpers.Request(client, http.MethodGet, "https://go-pjatk-server.fly.dev/api/game", nil, token, 5)
		if err2 != nil {
			return fmt.Errorf("cannot create request: %w", err2)
		}

		err = json.NewDecoder(resp2.Body).Decode(&data)
		if err != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}

	}

	var bo Board
	if coord == nil {
		resp4, err4 := helpers.Request(client, http.MethodGet, "https://go-pjatk-server.fly.dev/api/game/board", nil, token, 5)
		if err4 != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}

		err = json.NewDecoder(resp4.Body).Decode(&bo)
		if err != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}
	} else {
		bo.Board = coord
	}

	resp5, err6 := helpers.Request(client, http.MethodGet, "https://go-pjatk-server.fly.dev/api/game/desc", nil, token, 5)
	if err6 != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}

	var gDesc GameDescription
	err = json.NewDecoder(resp5.Body).Decode(&gDesc)
	if err != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}

	ui := gui.NewGUI(true)

	txt := gui.NewText(1, 1, "Welcome to the game Ships-GO :)", nil)
	ui.Draw(txt)

	sunk_ships := [4]int{4, 3, 2, 1}
	var ships_texts [4]*gui.Text

	ui.Draw(gui.NewText(70, 1, "Ships left to sink:", nil))
	for i := 0; i < len(sunk_ships); i++ {
		ships_texts[i] = gui.NewText(70, 2+i, (strconv.FormatInt(int64(i+1), 10) + " Tier Ship -> " + strconv.FormatInt(int64(sunk_ships[i]), 10)), nil)
	}

	for i := 0; i < len(sunk_ships); i++ {
		ui.Draw(ships_texts[i])
	}

	timer := gui.NewText(50, 2, "TIMER: ", nil)
	ui.Draw(timer)
	ui.Draw(gui.NewText(1, 2, "Press Ctrl+C to exit", nil))
	ui.Draw(gui.NewText(1, 3, data.Nick+" vs "+data.Opponent, nil))
	*name = data.Nick
	accu := gui.NewText(1, 4, "Shot precision: 0.0%", nil)
	ui.Draw(accu)
	if len(gDesc.Desc) > 90 {
		ui.Draw(gui.NewText(1, 29, gDesc.Desc[:45], nil))
		ui.Draw(gui.NewText(1, 30, gDesc.Desc[45:90]+"...", nil))
	} else if len(gDesc.Desc) > 45 {
		ui.Draw(gui.NewText(1, 29, gDesc.Desc[:45], nil))
		ui.Draw(gui.NewText(1, 30, gDesc.Desc[45:], nil))
	} else {
		ui.Draw(gui.NewText(1, 29, gDesc.Desc, nil))
	}

	if len(gDesc.Opp_desc) > 90 {
		ui.Draw(gui.NewText(50, 29, gDesc.Opp_desc[:45], nil))
		ui.Draw(gui.NewText(50, 30, gDesc.Opp_desc[45:90]+"...", nil))
	} else if len(gDesc.Opp_desc) > 45 {
		ui.Draw(gui.NewText(50, 29, gDesc.Opp_desc[:45], nil))
		ui.Draw(gui.NewText(50, 30, gDesc.Opp_desc[45:], nil))
	} else {
		ui.Draw(gui.NewText(50, 29, gDesc.Opp_desc, nil))
	}

	// boardConfig := gui.BoardConfig{RulerColor: gui.Color{Red: 236, Green: 54, Blue: 54}, TextColor: gui.Color{Red: 88, Green: 243, Blue: 212}}
	board := gui.NewBoard(1, 7, nil)
	board2 := gui.NewBoard(50, 7, nil)
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

	myHitCounter := 0.0
	myHits := 0.0
	oppHitCounter := 0
	coordsChecked := 0

	if len(data.Opp_shots) != 0 {

		shotsCounter := 0
		for i := coordsChecked; i < len(data.Opp_shots); i++ {
			x, y := ChangeCooerdinate(data.Opp_shots[i])
			shotsCounter++
			if states[x][y-1] == gui.Ship {
				oppHitCounter++
				states[x][y-1] = gui.Hit
				time.Sleep(time.Second * 1)
				Get_Opp_Shot(board, board2, &states, &states2, ui, &client, token, &oppHitCounter, &coordsChecked)
			} else {
				states[x][y-1] = gui.Miss
			}
			board.SetStates(states)
		}
		coordsChecked += shotsCounter
	}

	ctx, stop := context.WithCancel(context.Background())

	go func() {
		for {
			// time.Sleep(time.Millisecond * 300)
			data, err := GetGameStatus(&client, token)
			if err != nil {
				ui.Log("cannot get data: %w", err)
				break
			}

			if data.Game_status == "ended" {
				if data.Last_game_status == "win" {
					txt.SetText(fmt.Sprintf("YOU WON"))
				} else {
					txt.SetText(fmt.Sprintf("OPPONENT WON"))
					for _, coordinate := range bo.Board {
						x, y := ChangeCooerdinate(coordinate)
						if states[x][y-1] == gui.Ship {
							states[x][y-1] = gui.Hit
						}
					}
					board.SetStates(states)
				}
				time.Sleep(time.Second * 3)
				stop()
				break
			}

			if data.Should_fire {
				txt.SetText("Your turn!")

				ctx_ticker, stop_ticker := context.WithCancel(context.Background())
				ticker := time.NewTicker(time.Second)
				done := make(chan bool)

				go func() {
					tmp := data.Timer
					for {
						select {
						case <-done:
							ticker.Stop()
							return
						case <-ticker.C:
							tmp = tmp - 1
							timer.SetText("TIME: " + strconv.Itoa(tmp))
							if tmp <= 0 {
								stop_ticker()
							}
						}
					}
				}()

				Shot(&ctx_ticker, board, board2, &states, &states2, ui, &client, token, &myHitCounter, &myHits, accu, txt, &sunk_ships, ships_texts)
				timer.SetText("TIME: ")
				ticker.Stop()
				done <- true
			} else {
				txt.SetText("Opponent turn!")
				Get_Opp_Shot(board, board2, &states, &states2, ui, &client, token, &oppHitCounter, &coordsChecked)
			}
		}
	}()
	ui.Start(ctx, nil)

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

func Shot(ctx *context.Context, myBoard, oppBoard *gui.Board, myStates, oppStates *[10][10]gui.State, ui *gui.GUI, client *http.Client, token string, myHitCounter, myHits *float64, text, main *gui.Text, sunk_ships *[4]int, ships_texts [4]*gui.Text) error {
	var x, y int
	var char string
	bad_choice := false
	for !bad_choice {
		char = oppBoard.Listen(*ctx)
		if len(char) < 2 {
			return nil
		}
		ui.Log("My Shot: %s", char)
		x, y = ChangeCooerdinate(char)
		if oppStates[x][y-1] == gui.Hit || oppStates[x][y-1] == gui.Ship || oppStates[x][y-1] == gui.Miss {
			main.SetText("Field " + char + " is already marked!")
		} else {
			bad_choice = true
		}
	}

	*myHits++
	fire_data := Fire{Coord: char}
	encoded_data, _ := json.Marshal(fire_data)

	resp, err := helpers.Request(*client, http.MethodPost, "https://go-pjatk-server.fly.dev/api/game/fire", bytes.NewBuffer(encoded_data), token, 5)
	if err != nil {
		ui.Log("cannot send data: %w", err)
		return err
	}

	var data Fire_result

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		ui.Log("cannot unmarshal data: %w", err)
		return err
	}
	if data.Result == "miss" {
		oppStates[x][y-1] = gui.Miss
		oppBoard.SetStates(*oppStates)
	} else {
		oppStates[x][y-1] = gui.Hit
		*myHitCounter++
		if data.Result == "sunk" {
			var ship_tier *int
			value := 0
			ship_tier = &value
			SunkShip(oppStates, x, y, ship_tier)

			sunk_ships[*ship_tier-1] = sunk_ships[*ship_tier-1] - 1
			ships_texts[*ship_tier-1].SetText(strconv.FormatInt(int64(*ship_tier), 10) + " Tier Ship -> " + strconv.FormatInt(int64(sunk_ships[*ship_tier-1]), 10))
			ui.Draw(ships_texts[*ship_tier-1])
		}
		if *myHitCounter == 20 {
			return nil
		}
	}
	number := (*myHitCounter / (*myHits)) * 100
	tmp := fmt.Sprintf("%f", number)[:4]
	text.SetText("Shot precision: " + tmp + "%")
	oppBoard.SetStates(*oppStates)
	return nil
}

func Get_Opp_Shot(myBoard, oppBoard *gui.Board, myStates, oppStates *[10][10]gui.State, ui *gui.GUI, client *http.Client, token string, oppHitCounter, coordsChecked *int) error {
	time.Sleep(time.Millisecond * 300)
	data, err := GetGameStatus(client, token)
	if err != nil {
		ui.Log("cannot get data: %w", err)
		return err
	}

	ui.Log("DATA: %s", data)
	shotsCounter := 0
	if *coordsChecked == len(data.Opp_shots) {
		time.Sleep(time.Millisecond * 200)

		if data.Game_status == "ended" {
			return nil
		}
		Get_Opp_Shot(myBoard, oppBoard, myStates, oppStates, ui, client, token, oppHitCounter, coordsChecked)
	}
	for i := *coordsChecked; i < len(data.Opp_shots); i++ {

		x, y := ChangeCooerdinate(data.Opp_shots[i])
		shotsCounter++
		if myStates[x][y-1] == gui.Ship {
			myStates[x][y-1] = gui.Hit
			*oppHitCounter++
			if *oppHitCounter == 20 {
				return nil
			}
		} else {
			myStates[x][y-1] = gui.Miss
		}
		myBoard.SetStates(*myStates)
	}
	*coordsChecked += shotsCounter
	return nil
}

func GetGameStatus(client *http.Client, token string) (GameStatus, error) {
	var data GameStatus
	resp, err := helpers.Request(*client, http.MethodGet, "https://go-pjatk-server.fly.dev/api/game", nil, token, 5)
	if err != nil {
		return GameStatus{}, err
	}

	err3 := json.NewDecoder(resp.Body).Decode(&data)
	if err3 != nil {
		return GameStatus{}, err3
	}
	return data, nil
}

func ShortenDesc(desc string) string {
	ret := ""
	for i, l := range strings.Split(desc, "") {
		if i%21 == 0 {
			ret += "\t"
		}
		ret += l
	}
	return ret
}

func SunkShip(states *[10][10]gui.State, x, y int, counter *int) {
	if counter != nil {
		*counter++
	}
	states[x][y-1] = gui.Ship
	if x <= 9 && x >= 0 && y <= 9 && y >= 0 {
		if states[x][y] == gui.Hit {
			SunkShip(states, x, y+1, counter)
		} else {
			if states[x][y] != gui.Ship {
				states[x][y] = gui.Miss
			}
		}
	}
	if x <= 9 && x >= 0 && y-2 <= 9 && y-2 >= 0 {
		if states[x][y-2] == gui.Hit {
			SunkShip(states, x, y-1, counter)
		} else {
			if states[x][y-2] != gui.Ship {
				states[x][y-2] = gui.Miss
			}
		}
	}
	if x-1 <= 9 && x-1 >= 0 && y-1 <= 9 && y-1 >= 0 {
		if states[x-1][y-1] == gui.Hit {
			SunkShip(states, x-1, y, counter)
		} else {
			if states[x-1][y-1] != gui.Ship {
				states[x-1][y-1] = gui.Miss
			}
		}
	}
	if x+1 <= 9 && x+1 >= 0 && y-1 <= 9 && y-1 >= 0 {
		if states[x+1][y-1] == gui.Hit {
			SunkShip(states, x+1, y, counter)
		} else {
			if states[x+1][y-1] != gui.Ship {
				states[x+1][y-1] = gui.Miss
			}
		}
	}

	if x+1 <= 9 && x+1 >= 0 && y <= 9 && y >= 0 && states[x+1][y] != gui.Hit && states[x+1][y] != gui.Ship {
		states[x+1][y] = gui.Miss
	}

	if x-1 <= 9 && x-1 >= 0 && y <= 9 && y >= 0 && states[x-1][y] != gui.Hit && states[x-1][y] != gui.Ship {
		states[x-1][y] = gui.Miss
	}

	if x-1 <= 9 && x-1 >= 0 && y-2 <= 9 && y-2 >= 0 && states[x-1][y-2] != gui.Hit && states[x-1][y-2] != gui.Ship {
		states[x-1][y-2] = gui.Miss
	}

	if x+1 <= 9 && x+1 >= 0 && y-2 <= 9 && y-2 >= 0 && states[x+1][y-2] != gui.Hit && states[x+1][y-2] != gui.Ship {
		states[x+1][y-2] = gui.Miss
	}

	return
}
