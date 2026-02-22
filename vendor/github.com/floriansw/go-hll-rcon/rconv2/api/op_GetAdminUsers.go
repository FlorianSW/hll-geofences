package api

type GetAdminUsers struct {
}

type GetAdminUsersResponse struct {
	AdminUsers []AdminUserEntry `json:"AdminUsers"`
}

type AdminUserEntry struct {
	Id      string `json:"userId"`
	Group   string `json:"group"`
	Comment string `json:"comment"`
}
