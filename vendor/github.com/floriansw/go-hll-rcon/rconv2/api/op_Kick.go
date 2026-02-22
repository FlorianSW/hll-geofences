package api

type KickPlayer struct {
	Reason   string `json:"Reason"`
	PlayerId string `json:"PlayerId"`
}
