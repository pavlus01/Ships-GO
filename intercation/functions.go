package intercation

import (
	"context"
	"fmt"
	"shipsgo/game"
	"strconv"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

func SetBoard(myStates *[10][10]gui.State) []string {

	var coord []string
	ui := gui.NewGUI(true)
	ctx, stop := context.WithCancel(context.Background())

	txt := gui.NewText(1, 1, "Press on any coordinate to log it.", nil)
	ui.Draw(txt)
	ui.Draw(gui.NewText(1, 2, "Press Ctrl+C to exit", nil))

	sunk_ships := &[4]int{4, 3, 2, 1}
	var ships_texts [4]*gui.Text

	ui.Draw(gui.NewText(60, 10, "Click on red buttons to chose your fleet", nil))
	ui.Draw(gui.NewText(60, 12, "Now choosing:", nil))
	for i := 0; i < len(sunk_ships); i++ {
		ships_texts[i] = gui.NewText(60, 13+i, (strconv.FormatInt(int64(i+1), 10) + " Tier Ship -> " + strconv.FormatInt(int64(sunk_ships[i]), 10)), nil)
	}

	for i := 0; i < len(sunk_ships); i++ {
		ui.Draw(ships_texts[i])
	}

	board := gui.NewBoard(1, 4, nil)
	ui.Draw(board)

	go func() {
		counter := 0
		tmp := [10][10]gui.State{}
		go_back_n := [10][10]gui.State{}
		go_back_tmp := [10][10]gui.State{}
		var go_back_coord []string

	mainloop:
		for {

			for i := 0; i < len(sunk_ships); i++ {
				ui.Draw(ships_texts[i])
				if i == 4-counter {
					ships_texts[i].SetBgColor(gui.Green)
				} else {
					ships_texts[i].SetBgColor(gui.Grey)
				}
			}

			go_back_n = *myStates
			go_back_tmp = tmp
			go_back_coord = coord

			for f := 0; f < counter; f++ {

				for i := 0; i < 10; i++ {
					for j := 0; j < 10; j++ {
						if myStates[i][j] == gui.Hit {
							myStates[i][j] = gui.Empty
						} else if myStates[i][j] != gui.Ship && myStates[i][j] != gui.Empty {
							myStates[i][j] = gui.Hit
						}
					}
				}

				board.SetStates(*myStates)
				var x, y int

				bad_choice := false
				var char string
				for !bad_choice {
					char = board.Listen(context.TODO())
					txt.SetText(fmt.Sprintf("Chosen: %s", char))
					x, y = game.ChangeCooerdinate(char)
					if myStates[x][y-1] != gui.Hit {
						txt.SetText("Field " + char + " is not valid!")
					} else {
						myStates[x][y-1] = gui.Ship
						tmp[x][y-1] = gui.Hit
						bad_choice = true
						board.SetStates(*myStates)
					}
				}

				for i := 0; i < 10; i++ {
					for j := 0; j < 10; j++ {
						if myStates[i][j] != gui.Ship && myStates[i][j] != gui.Empty {
							myStates[i][j] = gui.Miss
						}
					}
				}

				RoundShip(myStates, &tmp, x, y)

				if counter != 4 && !EmptyCheck(myStates, x, y) {
					*myStates = go_back_n
					tmp = go_back_tmp
					coord = go_back_coord
					board.SetStates(*myStates)
					txt.SetText("Invalid operation, strat putting this thier ships once again!")
					continue mainloop
				}
				coord = append(coord, char)
				board.SetStates(*myStates)

				for i := 0; i < 4-counter; i++ {
					var char string
					bad_choice := false
					for !bad_choice {
						char = board.Listen(context.TODO())
						txt.SetText(fmt.Sprintf("Chosen: %s", char))
						x, y = game.ChangeCooerdinate(char)
						if myStates[x][y-1] != gui.Hit {
							txt.SetText("Field " + char + " is not valid!")
						} else {
							myStates[x][y-1] = gui.Ship
							tmp[x][y-1] = gui.Hit
							bad_choice = true
							board.SetStates(*myStates)
						}
					}
					RoundShip(myStates, &tmp, x, y)

					if i != (3-counter) && !EmptyCheck(myStates, x, y) {
						*myStates = go_back_n
						tmp = go_back_tmp
						coord = go_back_coord
						board.SetStates(*myStates)
						txt.SetText("Invalid operation, strat putting this thier ships once again!")
						continue mainloop
					}
					coord = append(coord, char)
					board.SetStates(*myStates)
				}

				var t *int
				game.SunkShip(&tmp, x, y, t)
			}
			if counter == 4 {
				break
			}
			counter++
		}

		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				if myStates[i][j] == gui.Hit || myStates[i][j] == gui.Miss {
					myStates[i][j] = gui.Empty
				}
			}
		}
		board.SetStates(*myStates)
		txt.SetText("That looks fantastic :)")
		time.Sleep(time.Second * 3)
		stop()
	}()

	ui.Start(ctx, nil)
	if len(coord) < 20 {
		return nil
	}
	return coord
}

func RoundShip(states, tmp *[10][10]gui.State, x, y int) {
	states[x][y-1] = gui.Ship
	if x <= 9 && x >= 0 && y <= 9 && y >= 0 {
		if states[x][y] != gui.Ship && tmp[x][y] != gui.Ship && tmp[x][y] != gui.Miss {
			states[x][y] = gui.Hit
		}
	}
	if x <= 9 && x >= 0 && y-2 <= 9 && y-2 >= 0 {
		if states[x][y-2] != gui.Ship && tmp[x][y-2] != gui.Ship && tmp[x][y-2] != gui.Miss {
			states[x][y-2] = gui.Hit
		}
	}
	if x-1 <= 9 && x-1 >= 0 && y-1 <= 9 && y-1 >= 0 {
		if states[x-1][y-1] != gui.Ship && tmp[x-1][y-1] != gui.Ship && tmp[x-1][y-1] != gui.Miss {
			states[x-1][y-1] = gui.Hit

		}
	}
	if x+1 <= 9 && x+1 >= 0 && y-1 <= 9 && y-1 >= 0 {
		if states[x+1][y-1] != gui.Ship && tmp[x+1][y-1] != gui.Ship && tmp[x+1][y-1] != gui.Miss {
			states[x+1][y-1] = gui.Hit
		}
	}
	if x+1 <= 9 && x+1 >= 0 && y <= 9 && y >= 0 && states[x+1][y] != gui.Hit && states[x+1][y] != gui.Ship {
		states[x+1][y] = gui.Empty
	}

	if x-1 <= 9 && x-1 >= 0 && y <= 9 && y >= 0 && states[x-1][y] != gui.Hit && states[x-1][y] != gui.Ship {
		states[x-1][y] = gui.Empty
	}

	if x-1 <= 9 && x-1 >= 0 && y-2 <= 9 && y-2 >= 0 && states[x-1][y-2] != gui.Hit && states[x-1][y-2] != gui.Ship {
		states[x-1][y-2] = gui.Empty
	}

	if x+1 <= 9 && x+1 >= 0 && y-2 <= 9 && y-2 >= 0 && states[x+1][y-2] != gui.Hit && states[x+1][y-2] != gui.Ship {
		states[x+1][y-2] = gui.Empty
	}

	return
}

func EmptyCheck(states *[10][10]gui.State, x, y int) bool {
	if x <= 9 && x >= 0 && y <= 9 && y >= 0 {
		if states[x][y] == gui.Hit {
			return true
		}
	}
	if x <= 9 && x >= 0 && y-2 <= 9 && y-2 >= 0 {
		if states[x][y-2] == gui.Hit {
			return true
		}
	}
	if x-1 <= 9 && x-1 >= 0 && y-1 <= 9 && y-1 >= 0 {
		if states[x-1][y-1] == gui.Hit {
			return true

		}
	}
	if x+1 <= 9 && x+1 >= 0 && y-1 <= 9 && y-1 >= 0 {
		if states[x+1][y-1] == gui.Hit {
			return true
		}
	}
	return false
}
