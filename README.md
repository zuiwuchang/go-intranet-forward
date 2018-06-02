# go-intranet-forward
使用 golang 實現的 內網穿透端口映射工具

# what
go-intranet-forward 以服務器方式 運行在一個 公網上 之後 以客戶端 方式 運行在一個 內網中 這時 go-intranet-forward 會把 內網的端口 映射到 公網 你就可以 通過 公網 連接到 內部網路 

go-intranet-forward 可以以 -s 參數 作為 服務器 運行 或 -c 參數 作為 客戶端 運行

要將一個 內網 端口 映射到 公網 需要 在 一個公網上 以 -s 運行 之後 以 -c 在內網 運行

# install
go-intranet-forward 使用了 [Protocol Buffers](https://developers.google.com/protocol-buffers/) 你需要 自己 配置好 golang 版本的 Protocol Buffers 環境
1. go get -u -d github.com/zuiwuchang/go-intranet-forward
2. cd ${GOPATH}/github.com/zuiwuchang/go-intranet-forward
3. build-proto.sh
4. build-go.sh
5. go install

# go-intranet-forward -s
go-intranet-forward -s server.jsonnet 將運行一個 映射 服務器 server.jsonnet 指定了 服務器配置 檔案
```jsonnet
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

            // 每次 recv 緩存 最大尺寸
            TunnelRecvBuffer:DefaultBuffer,
            // 每次 send 數據 最大尺寸
            TunnelSendBuffer:DefaultBuffer,
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
```

Server.Addr 指定的 服務器的 工作地址 go-intranet-forward -c 需要連接此地址 和 -s 服務器 完成通信

Forward 是一個數組 每條記錄 可以配置一個 映射服務 . 對於每個 映射 需要指定 一個 唯一的 標識 ID . Public 是映射後 開放的 公網 地址 任何 連接到此地址 的請求 都有 被自動 轉接到 內網

# go-intranet-forward -c
go-intranet-forward -c client.jsonnet 將運行一個 映射 客戶端 client.jsonnet 指定了 客戶端配置 檔案, 她將 自動 向 -s 服務器 完成 註冊 並建立好 端口 映射所需的 通信隧道
```jsonnet
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
```

Forward 是要向 服務器 註冊的 映射 服務 ,ID 指定了要註冊的 服務編號, Remote 指定了 服務器的 工作地址 應該是 服務器配置中的 Server.Addr ,Local 指定了 要將 區域網路中的 哪個地址 進行映射 這個地址 應該是 -c 客戶端 能夠訪問到的 任意地址
