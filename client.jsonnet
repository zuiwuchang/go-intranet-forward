local Second = 1;
local Minute = Second * 60;
local MinBuffer = 1024;
local DefaultBuffer = MinBuffer * 16;
{
    // 配置 轉發
    Forward:[
        {
            // 服務編號
            ID:1,
            // 遠端地址
            Remote:"192.168.16.3:9090",
            // 本機地址
            Local:"127.0.0.1:22",
            Key:"加密密鑰",
            Password:"連接密碼 如果為空 不驗證",

            // 每次 recv 緩存 最大尺寸
            RecvBuffer:DefaultBuffer,
            // 每次 send 數據 最大尺寸
            SendBuffer:DefaultBuffer,

            // 隧道 每次 recv 緩存 最大尺寸
            TunnelRecvBuffer:DefaultBuffer,
            // 隧道 每次 send 數據 最大尺寸
            TunnelSendBuffer:DefaultBuffer,

            // 重建 最長等待時間
	        MaxReconstruction:Minute,
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