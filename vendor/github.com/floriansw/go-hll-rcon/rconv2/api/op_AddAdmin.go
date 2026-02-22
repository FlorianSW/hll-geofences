package api

type AddAdmin struct {
	PlayerId   string `json:"PlayerId"`
	AdminGroup string `json:"AdminGroup"`
	Comment    string `json:"Comment"`
}
