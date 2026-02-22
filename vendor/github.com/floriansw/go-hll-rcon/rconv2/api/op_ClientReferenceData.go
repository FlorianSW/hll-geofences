package api

type GetClientReferenceData string

type GetClientReferenceDataResponse struct {
	Name        string      `json:"name"`
	Text        string      `json:"text"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"dialogueParameters"`
}

type Parameter struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	Id            string `json:"iD"`
	DisplayMember string `json:"displayMember"`
	ValueMember   string `json:"valueMember"`
}
