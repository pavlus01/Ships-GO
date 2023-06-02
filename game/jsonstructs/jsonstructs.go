package jsonstructs

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
