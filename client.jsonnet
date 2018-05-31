{
    // 配置 轉發
    Forwards:{
        // 遠端地址
        Remote:":9090",
        // 本機地址
        Local:"127.0.0.1:22",
        Key:"加密密鑰",
        Password:"連接密碼",
    },
    // 日誌 配置
    Log:{
        // 需要打印的 日誌等級
        Logs:[
            "trace",
            "debug",
            "info",
            "warn",
            "error",
            "fault"
        ],
        //是否 顯示 短檔案夾
        Short:true,
        // 是否 顯示 長檔案夾
        Long:false,
    },
}