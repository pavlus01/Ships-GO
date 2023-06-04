package intercation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"shipsgo/httphelper"
	"shipsgo/intercation/jsonstructs"
	"sort"
	"strconv"

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
	resp, err := httphelper.Request(client, http.MethodGet, "https://go-pjatk-server.fly.dev/api/game/list", nil, "", 5)
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}

	var data jsonstructs.Player_List

	fmt.Println("\nAvailable players list:\n")

	err = json.NewDecoder(resp.Body).Decode(&data)
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
