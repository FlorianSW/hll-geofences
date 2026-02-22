package api

type SetWarmupTimer struct {
	GameMode     string `json:"GameMode"`
	WarmupLength int32  `json:"WarmupLength"`
}
