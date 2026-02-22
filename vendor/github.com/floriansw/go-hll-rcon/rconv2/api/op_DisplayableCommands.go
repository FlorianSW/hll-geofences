package api

type GetDisplayableCommands struct {
}

type GetDisplayableCommandsResponse struct {
	Entries []DisplayableCommandEntry `json:"entries"`
}

type DisplayableCommandEntry struct {
	Id                string `json:"iD"`
	FriendlyName      string `json:"friendlyName"`
	IsClientSupported bool   `json:"isClientSupported"`
}
