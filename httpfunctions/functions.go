package httpfunctions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
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

func FirstConnection(desc, name, opponent string) error {
	var request_data Request_data
	if name != "" {
		request_data = Request_data{Desc: desc, Nick: name, Target_nick: "", Wpbot: false}
	} else {
		request_data = Request_data{Target_nick: "", Wpbot: false}
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

	resp, err := http.Post("https://go-pjatk-server.fly.dev/api/game", "application/json", bytes.NewBuffer(encoded_data))
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}
	token := resp.Header.Get("x-auth-token")

	var data GameStatus
	client := &http.Client{Timeout: time.Second * 5}

	for data.Game_status != "game_in_progress" {

		time.Sleep(time.Second * 1)
		req, err2 := http.NewRequest(http.MethodGet, "https://go-pjatk-server.fly.dev/api/game", nil)
		req.Header.Set("x-auth-token", token)
		resp2, err3 := client.Do(req)
		if err2 != nil {
			return fmt.Errorf("cannot create request: %w", err2)
		}
		if err3 != nil {
			return fmt.Errorf("cannot create request: %w", err3)
		}

		err = json.NewDecoder(resp2.Body).Decode(&data)
		if err != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}

	}

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

	req3, err6 := http.NewRequest("GET", "https://go-pjatk-server.fly.dev/api/game/desc", nil)
	req3.Header.Add("x-auth-token", token)
	if err6 != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}
	resp5, err7 := client.Do(req3)
	if err7 != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}
	var gDesc GameDescription
	err = json.NewDecoder(resp5.Body).Decode(&gDesc)
	if err != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}

	ui := gui.NewGUI(true)

	txt := gui.NewText(1, 1, "Press on any coordinate to log it.", nil)
	ui.Draw(txt)
	ui.Draw(gui.NewText(1, 2, "Press Ctrl+C to exit", nil))
	ui.Draw(gui.NewText(1, 3, data.Nick+" vs "+data.Opponent, nil))
	accu := gui.NewText(1, 4, "Shot precision: 0%", nil)
	ui.Draw(accu)
	if len(gDesc.Desc) > 90 {
		ui.Draw(gui.NewText(1, 28, gDesc.Desc[:45], nil))
		ui.Draw(gui.NewText(1, 29, gDesc.Desc[45:90]+"...", nil))
	} else if len(gDesc.Desc) > 45 {
		ui.Draw(gui.NewText(1, 28, gDesc.Desc[:45], nil))
		ui.Draw(gui.NewText(1, 29, gDesc.Desc[45:], nil))
	} else {
		ui.Draw(gui.NewText(1, 28, gDesc.Desc, nil))
	}

	if len(gDesc.Opp_desc) > 90 {
		ui.Draw(gui.NewText(50, 28, gDesc.Opp_desc[:45], nil))
		ui.Draw(gui.NewText(50, 29, gDesc.Opp_desc[45:90]+"...", nil))
	} else if len(gDesc.Opp_desc) > 45 {
		ui.Draw(gui.NewText(50, 28, gDesc.Opp_desc[:45], nil))
		ui.Draw(gui.NewText(50, 29, gDesc.Opp_desc[45:], nil))
	} else {
		ui.Draw(gui.NewText(50, 28, gDesc.Opp_desc, nil))
	}

	// boardConfig := gui.BoardConfig{RulerColor: gui.Color{Red: 236, Green: 54, Blue: 54}, TextColor: gui.Color{Red: 88, Green: 243, Blue: 212}}
	board := gui.NewBoard(1, 6, nil)
	board2 := gui.NewBoard(50, 6, nil)
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
	myHitCounter := 0.0
	myHits := 0.0
	oppHitCounter := 0
	coordsChecked := 0

	if len(data.Opp_shots) != 0 {

		ui.Log("DATA: %s", data)
		shotsCounter := 0
		for i := coordsChecked; i < len(data.Opp_shots); i++ {
			x, y := ChangeCooerdinate(data.Opp_shots[i])
			shotsCounter++
			if states[x][y-1] == gui.Ship {
				oppHitCounter++
				states[x][y-1] = gui.Hit
				time.Sleep(time.Second * 1)
				Get_Opp_Shot(board, board2, &states, &states2, ui, client, token, &oppHitCounter, &coordsChecked)
			} else {
				states[x][y-1] = gui.Miss
			}
			board.SetStates(states)
		}
		coordsChecked += shotsCounter
	}

	go func() {
		for {
			time.Sleep(time.Millisecond * 300)
			data, err := GetGameStatus(client, token)
			if err != nil {
				ui.Log("cannot get data: %w", err)
				break
			}

			if data.Game_status == "ended" {
				if data.Last_game_status == "win" {
					txt.SetText(fmt.Sprintf("YOU WON"))
					break
				} else {
					txt.SetText(fmt.Sprintf("OPPONENT WON"))
					for _, coordinate := range bo.Board {
						x, y := ChangeCooerdinate(coordinate)
						if states[x][y-1] == gui.Ship {
							states[x][y-1] = gui.Hit
						}
					}
					board.SetStates(states)
					break
				}
			}

			if data.Should_fire {
				Shot(board, board2, &states, &states2, ui, client, token, &myHitCounter, &myHits, accu)
			} else {
				Get_Opp_Shot(board, board2, &states, &states2, ui, client, token, &oppHitCounter, &coordsChecked)
			}
		}
	}()
	ctx := context.Background()
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

func Shot(myBoard, oppBoard *gui.Board, myStates, oppStates *[10][10]gui.State, ui *gui.GUI, client *http.Client, token string, myHitCounter, myHits *float64, text *gui.Text) error {
	var x, y int
	char := oppBoard.Listen(context.TODO())
	ui.Log("My Shot: %s", char)
	*myHits++
	x, y = ChangeCooerdinate(char)
	fire_data := Fire{Coord: char}
	encoded_data, _ := json.Marshal(fire_data)

	req, err := http.NewRequest("POST", "https://go-pjatk-server.fly.dev/api/game/fire", bytes.NewBuffer(encoded_data))
	req.Header.Add("x-auth-token", token)
	if err != nil {
		ui.Log("cannot send data: %w", err)
		return err
	}
	resp, err := client.Do(req)
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
	ui.Log("result: %s", data.Result)
	if data.Result == "miss" {
		oppStates[x][y-1] = gui.Miss
		oppBoard.SetStates(*oppStates)
	} else {
		oppStates[x][y-1] = gui.Hit
		*myHitCounter++
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
	time.Sleep(time.Millisecond * 500)
	data, err := GetGameStatus(client, token)
	if err != nil {
		ui.Log("cannot get data: %w", err)
		return err
	}

	ui.Log("DATA: %s", data)
	shotsCounter := 0
	if *coordsChecked == len(data.Opp_shots) {
		time.Sleep(time.Millisecond * 200)
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
	req, err := http.NewRequest(http.MethodGet, "https://go-pjatk-server.fly.dev/api/game", nil)
	req.Header.Set("x-auth-token", token)
	resp, err2 := client.Do(req)
	if err != nil || err2 != nil {
		if err != nil {
			return GameStatus{}, err
		}
		if err2 != nil {
			return GameStatus{}, err2
		}
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
