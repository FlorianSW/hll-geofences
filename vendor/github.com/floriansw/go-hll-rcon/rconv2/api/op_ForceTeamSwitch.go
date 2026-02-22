package api

type ForceMode uint8

const (
	ForceModeOnDeath ForceMode = iota
	ForceModeImmediately
)

type ForceTeamSwitch struct {
	ForceMode ForceMode `json:"ForceMode"`
	PlayerId  string    `json:"PlayerId"`
}
