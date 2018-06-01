package configure

// ClientForward .
type ClientForward struct {
	// 服務編號
	ID uint32 `json:"ID"`
	// 遠端地址
	Remote string
	// 本機地址
	Local string
	// 加密密鑰
	Key string
	// 連接密碼
	Password string

	// 每次 recv 緩存 最大尺寸
	RecvBuffer int
	// 每次 send 數據 最大尺寸
	SendBuffer int
	// 隧道 每次 recv 緩存 最大尺寸
	TunnelRecvBuffer int
	// 隧道 每次 send 數據 最大尺寸
	TunnelSendBuffer int
}
