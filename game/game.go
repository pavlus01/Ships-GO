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

type GuiFields struct {
	ui          *gui.GUI
	txt         *gui.Text
	sunk_ships  *[4]int
	ships_texts [4]*gui.Text
	timer       *gui.Text
	accu        *gui.Text
	board       *gui.Board
	board2      *gui.Board
	states      *[10][10]gui.State
	states2     *[10][10]gui.State
}

func Game(desc string, name *string, opponent string, client http.Client, coord []string) error {
	var request_data jsonstructs.Request_data
	if *name != "" {
		request_data = jsonstructs.Request_data{Coords: coord, Desc: desc, Nick: *name, Target_nick: "", Wpbot: false}
	} else {
		request_data = jsonstructs.Request_data{Coords: coord, Target_nick: "", Wpbot: false}
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

	resp, err := httphelper.Request(client, http.MethodPost, "https://go-pjatk-server.fly.dev/api/game", bytes.NewBuffer(encoded_data), "", 5)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}
	token := resp.Header.Get("x-auth-token")

	var data jsonstructs.GameStatus

	for data.Game_status != "game_in_progress" {

		time.Sleep(time.Second * 1)
		resp, err = httphelper.Request(client, http.MethodGet, "https://go-pjatk-server.fly.dev/api/game", nil, token, 5)
		if err != nil {
			return fmt.Errorf("cannot create request: %w", err)
		}

		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}

	}

	var bo jsonstructs.Board
	if coord == nil {
		resp, err = httphelper.Request(client, http.MethodGet, "https://go-pjatk-server.fly.dev/api/game/board", nil, token, 5)
		if err != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}

		err = json.NewDecoder(resp.Body).Decode(&bo)
		if err != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}
	} else {
		bo.Board = coord
	}

	resp, err = httphelper.Request(client, http.MethodGet, "https://go-pjatk-server.fly.dev/api/game/desc", nil, token, 5)
	if err != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}

	var gDesc jsonstructs.GameDescription
	err = json.NewDecoder(resp.Body).Decode(&gDesc)
	if err != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}

	gf := DrawGUI(bo, gDesc, data)
	*name = data.Nick

	myHitCounter := 0.0
	myHits := 0.0
	oppHitCounter := 0
	coordsChecked := 0

	ctx, stop := context.WithCancel(context.Background())

	go func() {
		for {
			// time.Sleep(time.Millisecond * 300)
			data, err := GetGameStatus(&client, token)
			if err != nil {
				gf.ui.Log("cannot get data: %w", err)
				break
			}

			if data.Game_status == "ended" {
				if data.Last_game_status == "win" {
					gf.txt.SetText("YOU WON")
					gf.txt.SetBgColor(gui.NewColor(55, 255, 180))
				} else {
					gf.txt.SetText("OPPONENT WON")
					gf.txt.SetBgColor(gui.NewColor(237, 16, 33))
					for _, coordinate := range bo.Board {
						x, y := ChangeCooerdinate(coordinate)
						if gf.states[x][y-1] == gui.Ship {
							gf.states[x][y-1] = gui.Hit
						}
					}
					gf.board.SetStates(*gf.states)
				}
				time.Sleep(time.Second * 3)
				stop()
				break
			}

			if data.Should_fire {
				gf.txt.SetText("Your turn!")

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
							gf.timer.SetText("TIME: " + strconv.Itoa(tmp))
							if tmp < 5 {
								gf.timer.SetBgColor(gui.Red)
							}
							if tmp <= 0 {
								stop_ticker()
							}
						}
					}
				}()

				Shot(&ctx_ticker, gf, &client, token, &myHitCounter, &myHits)
				gf.timer.SetText("TIME: ")
				ticker.Stop()
				done <- true
			} else {
				gf.txt.SetText("Opponent turn!")
				Get_Opp_Shot(gf, &client, token, &oppHitCounter, &coordsChecked)
			}
		}
	}()
	gf.ui.Start(ctx, nil)

	data, err = GetGameStatus(&client, token)
	if err != nil {
		gf.ui.Log("cannot get data: %w", err)
	}
	if data.Game_status == "game_in_progress" {
		httphelper.Request(client, http.MethodDelete, "https://go-pjatk-server.fly.dev/api/game/abandon", nil, token, 5)
	}

	return nil
}
