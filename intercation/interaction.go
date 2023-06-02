package intercation

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"shipsgo/game"
	"shipsgo/httphelper"
	"shipsgo/intercation/jsonstructs"
	"sort"
	"strconv"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

func PlayerDescription() (string, string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter nick(empty = random): ")
	nick, _ := reader.ReadString('\n')
	desc := ""
	if nick != "\n" {
		fmt.Print("Enter description: ")
		desc, _ = reader.ReadString('\n')
		desc = desc[:(len(desc) - 1)]
	}
	nick = nick[:(len(nick) - 1)]
	return nick, desc
}

func ShowPlayersList(client http.Client) (string, error) {
	resp2, err2 := httphelper.Request(client, http.MethodGet, "https://go-pjatk-server.fly.dev/api/game/list", nil, "", 5)
	if err2 != nil {
		return "", fmt.Errorf("cannot create request: %w", err2)
	}

	var data jsonstructs.Player_List

	fmt.Println("\nAvailable players list:\n")

	err := json.NewDecoder(resp2.Body).Decode(&data)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshal data: %w", err)
	}
	fmt.Println("0\twpbot\t\twaiting")
	for index := range data {
		fmt.Println(strconv.Itoa(index+1) + "\t" + data[index].Nick + "\t\t" + data[index].Game_status)
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter opponent id(empty = waiting mode): ")
	oppoId, _ := reader.ReadString('\n')
	oppoId = oppoId[:(len(oppoId) - 1)]
	if oppoId == "" {
		fmt.Println("\nWaiting for game challenge")
		return "", nil
	}
	Id, _ := strconv.ParseInt(oppoId, 10, 64)
	if Id > int64(len(data)) {
		return "", fmt.Errorf("given index is too high")
	}
	if Id == 0 {
		return "wpbot", nil
	}
	return data[Id-1].Nick, nil
}

func PostGameStatistics(nick *string, client http.Client) error {
	resp2, err2 := httphelper.Request(client, http.MethodGet, "https://go-pjatk-server.fly.dev/api/stats", nil, "", 5)
	if err2 != nil {
		return err2
	}

	var data jsonstructs.Player_Stats
	isIn := false

	fmt.Println("\nTOP 10 Players list:")
	fmt.Println("Nick\t\t\t\tGames\tWins\tPoints")
	fmt.Println("------------------------------------------------------------------")

	err := json.NewDecoder(resp2.Body).Decode(&data)
	if err != nil {
		return err
	}

	sort.Slice(data.Stats, func(i, j int) bool {
		return data.Stats[i].Wins > data.Stats[j].Wins
	})

	for index := range data.Stats {
		if len(data.Stats[index].Nick) < 8 {
			data.Stats[index].Nick += "\t"
		}

		if len(data.Stats[index].Nick) < 16 {
			data.Stats[index].Nick += "\t"
		}

		if data.Stats[index].Nick == *nick {
			isIn = true
		}
		fmt.Println(data.Stats[index].Nick + "\t\t" + strconv.Itoa(data.Stats[index].Games) + "\t" + strconv.Itoa(data.Stats[index].Wins) + "\t" + strconv.Itoa(data.Stats[index].Points))
	}
	if !isIn {
		fmt.Println("------------------------------------------------------------------")
		resp2, err2 := httphelper.Request(client, http.MethodGet, "https://go-pjatk-server.fly.dev/api/stats/"+*nick, nil, "", 5)
		if err2 != nil {
			return err2
		}

		var data2 jsonstructs.Your_Stats

		err := json.NewDecoder(resp2.Body).Decode(&data2)
		if err != nil {
			return err
		}

		if len(data2.Stats.Nick) < 8 {
			data2.Stats.Nick += "\t"
		}

		if len(data2.Stats.Nick) < 16 {
			data2.Stats.Nick += "\t"
		}

		fmt.Println(data2.Stats.Nick + "\t\t" + strconv.Itoa(data2.Stats.Games) + "\t" + strconv.Itoa(data2.Stats.Wins) + "\t" + strconv.Itoa(data2.Stats.Points))
	}
	return nil
}

func OwnBoard() []string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want put ships by yourself(empty = no): ")
	tmp, _ := reader.ReadString('\n')
	var coords []string
	if tmp != "\n" {
		states := [10][10]gui.State{}
		coords = SetBoard(&states)
	} else {
		coords = nil
	}
	return coords
}

func SetBoard(myStates *[10][10]gui.State) []string {

	var coord []string
	ui := gui.NewGUI(true)
	ctx, stop := context.WithCancel(context.Background())

	txt := gui.NewText(1, 1, "Press on any coordinate to log it.", nil)
	ui.Draw(txt)
	ui.Draw(gui.NewText(1, 2, "Press Ctrl+C to exit", nil))

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
