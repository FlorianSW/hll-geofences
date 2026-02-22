package api

type RemovePlayerFromPlatoon struct {
	PlayerId string `json:"PlayerId"`
	Reason   string `json:"Reason"`
}
