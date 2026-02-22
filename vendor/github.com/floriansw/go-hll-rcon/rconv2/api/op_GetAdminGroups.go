package api

type GetAdminGroups struct {
}

type GetAdminGroupsResponse struct {
	GroupNames []string `json:"GroupNames"`
}
