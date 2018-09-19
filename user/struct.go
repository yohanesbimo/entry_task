package user

type ResponseMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type UserProfile struct {
	ID       int64  `json:"id, omitempty"`
	Username string `json:"username"`
	Password string `json:"omitempty"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Photo    string `json:"photo"`
}
