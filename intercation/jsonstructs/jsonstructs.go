package jsonstructs


type Player_List []struct {
	Game_status string `json:"game_status"`
	Nick        string `json:"nick"`
}

type Player_Stats struct {
	Stats []struct {
		Games  int    `json:"games"`
		Nick   string `json:"nick"`
		Points int    `json:"points"`
		Wins   int    `json:"wins"`
	} `json:"stats"`
}

type Your_Stats struct {
	Stats struct {
		Games  int    `json:"games"`
		Nick   string `json:"nick"`
		Points int    `json:"points"`
		Wins   int    `json:"wins"`
	} `json:"stats"`
}