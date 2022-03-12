package kasa

type Response struct {
	ErrorCode int         `json:"error_code"`
	Result    interface{} `json:"result"`
}

// type Response struct {
// 	Data data `json:"data"`
// }

type LoginResponse struct {
	AccountId string `json:"accountId"`
	RegTime   string `json:"regTime"`
	Email     string `json:"email"`
	Token     string `json:"token"`
}

type LoginError struct {
	ErrorCode int
	Err       error
}

func (e *LoginError) Error() string {
	return e.Err.Error()
}
