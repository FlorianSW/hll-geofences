package api

type DisbandPlatoon struct {
	TeamIndex  int32  `json:"TeamIndex"`
	SquadIndex int32  `json:"SquadIndex"`
	Reason     string `json:"Reason"`
}
