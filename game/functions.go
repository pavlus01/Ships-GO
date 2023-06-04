package game

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"shipsgo/game/jsonstructs"
	"shipsgo/httphelper"
	"strconv"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

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

func Shot(ctx *context.Context, gf GuiFields, client *http.Client, token string, myHitCounter, myHits *float64) error {
	var x, y int
	var char string
	bad_choice := false
	for !bad_choice {
		char = gf.board2.Listen(*ctx)
		if len(char) < 2 {
			return nil
		}
		// gf.ui.Log("My Shot: %s", char)
		x, y = ChangeCooerdinate(char)
		if gf.states2[x][y-1] == gui.Hit || gf.states2[x][y-1] == gui.Ship || gf.states2[x][y-1] == gui.Miss {
			gf.txt.SetText("Field " + char + " is already marked!")
		} else {
			bad_choice = true
		}
	}

	*myHits++
	fire_data := jsonstructs.Fire{Coord: char}
	encoded_data, _ := json.Marshal(fire_data)

	resp, err := httphelper.Request(*client, http.MethodPost, "https://go-pjatk-server.fly.dev/api/game/fire", bytes.NewBuffer(encoded_data), token, 5)
	if err != nil {
		gf.ui.Log("cannot send data: %w", err)
		return err
	}

	var data jsonstructs.Fire_result

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		gf.ui.Log("cannot unmarshal data: %w", err)
		return err
	}
	if data.Result == "miss" {
		gf.states2[x][y-1] = gui.Miss
		gf.board2.SetStates(*gf.states2)
	} else {
		gf.states2[x][y-1] = gui.Hit
		*myHitCounter++
		if data.Result == "sunk" {
			value := 0
			ship_tier := &value
			SunkShip(gf.states2, x, y, ship_tier)

			gf.sunk_ships[*ship_tier-1] = gf.sunk_ships[*ship_tier-1] - 1
			gf.ships_texts[*ship_tier-1].SetText(strconv.FormatInt(int64(*ship_tier), 10) + " Tier Ship -> " + strconv.FormatInt(int64(gf.sunk_ships[*ship_tier-1]), 10))
			gf.ships_texts[*ship_tier-1].SetBgColor(gui.Color{Red: 161, Green: uint8(*ship_tier * 60), Blue: 54})
			gf.ui.Draw(gf.ships_texts[*ship_tier-1])
		}
		if *myHitCounter == 20 {
			return nil
		}
	}
	number := (*myHitCounter / (*myHits)) * 100
	tmp := fmt.Sprintf("%f", number)[:4]
	gf.accu.SetText("Shot precision: " + tmp + "%")
	gf.accu.SetBgColor(gui.Color{Red: 240 - uint8(240*number/100), Green: 88, Blue: 87})
	gf.board2.SetStates(*gf.states2)
	return nil
}

func Get_Opp_Shot(gf GuiFields, client *http.Client, token string, oppHitCounter, coordsChecked *int) error {
	time.Sleep(time.Millisecond * 300)
	data, err := GetGameStatus(client, token)
	if err != nil {
		gf.ui.Log("cannot get data: %w", err)
		return err
	}

	// gf.ui.Log("DATA: %s", data)
	shotsCounter := 0
	if *coordsChecked == len(data.Opp_shots) {
		time.Sleep(time.Millisecond * 200)

		if data.Game_status == "ended" {
			return nil
		}
		Get_Opp_Shot(gf, client, token, oppHitCounter, coordsChecked)
	}
	for i := *coordsChecked; i < len(data.Opp_shots); i++ {

		x, y := ChangeCooerdinate(data.Opp_shots[i])
		shotsCounter++
		if gf.states[x][y-1] == gui.Ship {
			gf.states[x][y-1] = gui.Hit
			*oppHitCounter++
			if *oppHitCounter == 20 {
				return nil
			}
		} else {
			gf.states[x][y-1] = gui.Miss
		}
		gf.board.SetStates(*gf.states)
	}
	*coordsChecked += shotsCounter
	return nil
}

func GetGameStatus(client *http.Client, token string) (jsonstructs.GameStatus, error) {
	var data jsonstructs.GameStatus
	resp, err := httphelper.Request(*client, http.MethodGet, "https://go-pjatk-server.fly.dev/api/game", nil, token, 5)
	if err != nil {
		return jsonstructs.GameStatus{}, err
	}

	err3 := json.NewDecoder(resp.Body).Decode(&data)
	if err3 != nil {
		return jsonstructs.GameStatus{}, err3
	}
	return data, nil
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

func DrawGUI(bo jsonstructs.Board, gDesc jsonstructs.GameDescription, data jsonstructs.GameStatus) GuiFields {
	var gf GuiFields
	gf.ui = gui.NewGUI(true)

	gf.txt = gui.NewText(1, 1, "Welcome to the game Ships-GO :)", nil)
	gf.ui.Draw(gf.txt)

	gf.sunk_ships = &[4]int{4, 3, 2, 1}

	gf.ui.Draw(gui.NewText(70, 1, "Ships left to sink:", nil))
	for i := 0; i < len(gf.sunk_ships); i++ {
		gf.ships_texts[i] = gui.NewText(70, 2+i, (strconv.FormatInt(int64(i+1), 10) + " Tier Ship -> " + strconv.FormatInt(int64(gf.sunk_ships[i]), 10)), nil)
	}

	for i := 0; i < len(gf.sunk_ships); i++ {
		gf.ships_texts[i].SetBgColor(gui.Color{Red: 161, Green: uint8(i * 60), Blue: 54})
		gf.ui.Draw(gf.ships_texts[i])
	}

	gf.timer = gui.NewText(50, 2, "TIMER: ", nil)
	gf.ui.Draw(gf.timer)
	gf.ui.Draw(gui.NewText(1, 2, "Press Ctrl+C to exit", nil))
	gf.ui.Draw(gui.NewText(1, 3, data.Nick+" vs "+data.Opponent, nil))
	gf.accu = gui.NewText(1, 4, "Shot precision: 0.0%", nil)
	gf.ui.Draw(gf.accu)
	if len(gDesc.Desc) > 90 {
		gf.ui.Draw(gui.NewText(1, 29, gDesc.Desc[:45], nil))
		gf.ui.Draw(gui.NewText(1, 30, gDesc.Desc[45:90]+"...", nil))
	} else if len(gDesc.Desc) > 45 {
		gf.ui.Draw(gui.NewText(1, 29, gDesc.Desc[:45], nil))
		gf.ui.Draw(gui.NewText(1, 30, gDesc.Desc[45:], nil))
	} else {
		gf.ui.Draw(gui.NewText(1, 29, gDesc.Desc, nil))
	}

	if len(gDesc.Opp_desc) > 90 {
		gf.ui.Draw(gui.NewText(50, 29, gDesc.Opp_desc[:45], nil))
		gf.ui.Draw(gui.NewText(50, 30, gDesc.Opp_desc[45:90]+"...", nil))
	} else if len(gDesc.Opp_desc) > 45 {
		gf.ui.Draw(gui.NewText(50, 29, gDesc.Opp_desc[:45], nil))
		gf.ui.Draw(gui.NewText(50, 30, gDesc.Opp_desc[45:], nil))
	} else {
		gf.ui.Draw(gui.NewText(50, 29, gDesc.Opp_desc, nil))
	}

	boardConfig := gui.BoardConfig{RulerColor: gui.Color{Red: 241, Green: 91, Blue: 54}, TextColor: gui.Color{Red: 105, Green: 30, Blue: 128},
		EmptyColor: gui.Grey, HitColor: gui.Red, MissColor: gui.Blue, ShipColor: gui.Green, EmptyChar: 'E', HitChar: 'H', MissChar: 'M', ShipChar: 'S'}
	gf.board = gui.NewBoard(1, 7, &boardConfig)
	gf.board2 = gui.NewBoard(50, 7, &boardConfig)
	gf.ui.Draw(gf.board)
	gf.ui.Draw(gf.board2)
	gf.states = &[10][10]gui.State{}
	gf.states2 = &[10][10]gui.State{}
	for i := range gf.states {
		gf.states[i] = [10]gui.State{}
	}

	for _, coordinate := range bo.Board {
		x, y := ChangeCooerdinate(coordinate)
		gf.states[x][y-1] = gui.Ship
	}
	gf.board.SetStates(*gf.states)

	return gf
}
