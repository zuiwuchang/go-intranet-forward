package configure

// Log 日誌 配置
type Log struct {
	// 需要打印的 日誌 等級
	Logs []string
	//是否 顯示 短檔案夾
	Short bool
	// 是否 顯示 長檔案夾
	Long bool
}
