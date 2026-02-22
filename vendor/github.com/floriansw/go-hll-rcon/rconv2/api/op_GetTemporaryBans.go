package api

type GetTemporaryBans struct {
}

type GetTemporaryBansResponse struct {
	BanList []BanListEntry `json:"banList"`
}
