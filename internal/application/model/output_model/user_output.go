package outputmodel

type UserOutput struct {
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	DepartmentName string `json:"departmentName"`
	Icon          string `json:"icon"`
	IsSupporter   bool   `json:"isSupporter"`
	IsAdmin       bool   `json:"isAdmin"`
}
