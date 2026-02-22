package api

type SetMatchTimer struct {
	GameMode    string `json:"GameMode"`
	MatchLength int32  `json:"MatchLength"`
}
