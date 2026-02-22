package api

type SetDynamicWeatherEnabled struct {
	MapId  string `json:"MapId"`
	Enable bool   `json:"Enable"`
}
