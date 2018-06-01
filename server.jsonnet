local Second = 1;
local Minute = Second * 60;
local MinInitTimeout = Second;
local DefaultInitTimeout = Minute;
local MinBuffer = 1024;
local DefaultBuffer = MinBuffer * 16;
{
    // 服務器 配置
    Server:{
        // listen 地址
        Addr:":9090",
        // 初始化超時時間
        InitTimeout:DefaultInitTimeout,

        // 每次 recv 緩存 最大尺寸
        RecvBuffer:DefaultBuffer,
        // 每次 send 數據 最大尺寸
        SendBuffer:DefaultBuffer,
    },
    // 配置 轉發
    Forward:[
        {
            // 服務編號
            ID:1,
            // 公網 地址
            Public:":10000",
            Key:"加密密鑰",
            Password:"連接密碼 如果為空 不驗證",
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