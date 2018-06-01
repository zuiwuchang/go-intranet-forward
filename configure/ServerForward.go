package configure

// ServerForward .
type ServerForward struct {
	// 服務編號
	ID uint32 `json:"ID"`
	// 公網 地址
	Public string
	// 加密密鑰
	Key string
	// 連接密碼 如果為空 不驗證
	Password string
}
