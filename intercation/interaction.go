package intercation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Player_List []struct {
	Game_status string `json:"game_status"`
	Nick        string `json:"nick"`
}

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

func ShowPlayersList() (string, error) {
	client := &http.Client{Timeout: time.Second * 5}
	req, err2 := http.NewRequest(http.MethodGet, "https://go-pjatk-server.fly.dev/api/game/list", nil)
	resp2, err3 := client.Do(req)
	if err2 != nil {
		return "", fmt.Errorf("cannot create request: %w", err2)
	}
	if err3 != nil {
		return "", fmt.Errorf("cannot create request: %w", err3)
	}

	var data Player_List

	fmt.Println("Available players list: ")

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
