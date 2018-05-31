local Second = 1000;
local Minute = Second * 60;
local Hour = Minute * 60;
{
    // 服務器 配置
    Server:{
        // listen 地址
        Addr:":9090",
        // 超時斷線 為0 永不超時
        Timeout:Hour,
    },
    // 配置 轉發
    Forward:[
        {
            // 服務編號
            ID:1,
            // 公網 地址
            Public:":10000",
            Key:"加密密鑰",
            Password:"連接密碼",
        },
    ],
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