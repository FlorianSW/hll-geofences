package api

import "time"

type GetPermanentBans struct {
}

type GetPermanentBansResponse struct {
	BanList []BanListEntry `json:"banList"`
}

type BanListEntry struct {
	Id        string    `json:"userId"`
	Name      string    `json:"userName"`
	Banned    time.Time `json:"timeOfBanning"`
	Duration  int       `json:"durationHours"`
	Reason    string    `json:"banReason"`
	AdminName string    `json:"adminName"`
}
