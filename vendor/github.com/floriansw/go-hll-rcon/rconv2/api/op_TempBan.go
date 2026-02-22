package api

type TemporaryBanPlayer struct {
	Reason    string `json:"Reason"`
	PlayerId  string `json:"PlayerId"`
	Duration  int32  `json:"Duration"`
	AdminName string `json:"AdminName"`
}
