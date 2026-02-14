# ApiAgent 使用文档

## 项目概述

ApiAgent 是一个**云手机集控平台的本地代理网关**，用于连接远程集控平台（`minio.accjs.cn`），通过 WebRTC 打洞技术与中间件/设备建立 P2P 连接，并在本地提供 HTTP/WebSocket 代理接口，实现对云手机设备的统一管理和控制。

### 核心能力

- 🔐 **集控平台认证** — 通过 API Key 登录集控平台，获取操作令牌
- 📱 **设备管理** — 获取设备列表、设备详情、在线状态
- 🌐 **WebRTC 打洞** — 与中间件/设备建立 P2P 数据通道
- 🔄 **HTTP 反向代理** — 拦截特定请求做本地处理，其余转发到集控平台
- 🖥️ **WebSocket 桥接** — 将 WebSocket 连接桥接到 RTC 中间件通道
- 🗺️ **端口映射** — 将远程设备端口映射到本地（支持 SOCKS5 代理模式）
- 📺 **H264 串流** — 获取设备视频/音频流

---

## 项目结构

```
ApiAgent/
├── app/                          # 主应用
│   ├── main.go                   # 入口文件
│   ├── go.mod
│   └── table/coreClass/
│       ├── centControlPlatform/  # 集控平台核心模块
│       │   ├── Type.go           # Core 结构体定义
│       │   ├── Function.go       # 登录、获取设备列表
│       │   ├── centGetToken.go   # 获取各类 RTC Token
│       │   ├── creatRtcObject.go # 创建 RTC 连接对象
│       │   ├── httpServer.go     # HTTP/WS 代理服务
│       │   └── httpFuc.go        # HTTP 业务处理函数
│       ├── middleAgentRtc/       # 中间件 RTC 连接管理
│       │   ├── rtc.go            # 连接建立、断开、数据处理
│       │   ├── syncSendMessage.go    # 同步消息（请求-响应）
│       │   ├── asycSendMessage.go    # 异步消息（批量操作）
│       │   └── eventMessages.go      # 事件处理（上下线、屏幕旋转）
│       ├── deviceRtc/            # 设备直连 RTC（H264 串流）
│       │   └── rtc.go
│       ├── portMapSocket5Rtc/    # 端口映射/SOCKS5 RTC
│       │   └── rtc.go
│       └── publicStruct/         # 公共数据结构
│           └── type.go
├── pkg/                          # 公共库
│   ├── netclient/                # 网络客户端封装（RTC/TCP/WS）
│   ├── socks5/                   # SOCKS5 代理实现
│   ├── portmap/                  # 端口转发器
│   ├── public/                   # 公共工具（配置、消息、加密）
│   └── utils/                    # 工具函数（压缩、加密、设备信息）
└── go.work.sum
```

---

## 架构说明

```
┌─────────────┐     HTTP/WS      ┌──────────────┐     WebRTC      ┌──────────────┐
│  本地客户端   │ ──────────────→ │   ApiAgent    │ ─────────────→ │   中间件      │
│  (浏览器等)   │  127.0.0.2:8888 │  (本地代理)    │   P2P 打洞     │  (管理设备)    │
└─────────────┘                  └──────────────┘                 └──────┬───────┘
                                       │                                 │
                                       │ HTTPS 转发                      │ 控制
                                       ▼                                 ▼
                                 ┌──────────────┐               ┌──────────────┐
                                 │  集控平台      │               │  云手机设备    │
                                 │ minio.accjs.cn│               │  (Android)    │
                                 └──────────────┘               └──────────────┘
```

### 三种 RTC 连接类型

| 类型 | 用途 | 连接目标 |
|------|------|----------|
| **中间件 RTC** | 设备管理、批量控制、截图 | 中间件（管理多台设备） |
| **设备 RTC** | H264 视频流、音频流 | 单台设备直连 |
| **端口映射 RTC** | 端口转发、SOCKS5 代理 | 单台设备直连 |

---

## 快速开始

### 1. 初始化

```go
import "app/table/coreClass/centControlPlatform"

// 创建核心对象（自动启动 HTTP 代理服务）
core := centControlPlatform.NewCore("minio.accjs.cn", "你的API_KEY")
```

`NewCore` 会自动：
- 初始化 HTTP 反向代理服务（监听 `127.0.0.2:8888`）
- 通过 WebSocket 连接集控平台
- 使用 API Key 登录获取 Token

### 2. 获取设备列表

```go
list, err := core.CenterGetDeviceList()
// list.Devices      — 所有设备
// list.MiddleWareIds — 所有中间件 ID
```

### 3. 连接中间件

```go
middleAgent, err := core.CreatMiddlewareRtc(
    middlewareId,   // 中间件 ID
    onOpen,         // 连接成功回调（可为 nil）
    onClose,        // 连接断开回调（可为 nil）
    onMessage,      // 收到消息回调（可为 nil）
)
```

### 4. 控制设备

连接中间件后，可通过同步或异步方式控制设备：

```go
// 同步方式 — 发送并等待响应
result, err := middleAgent.SyncGetDeviceDetails(deviceId)
result, err := middleAgent.SyncShellCommand(deviceId, "ls /sdcard")
result, err := middleAgent.SyncGetAppList(deviceId)
result, err := middleAgent.SyncRunApp(deviceId, "com.example.app")

// 异步方式 — 批量操作，通过回调获取结果
middleAgent.AsyncTouch(&touchReq, deviceIds)
middleAgent.AsyncKeyBoard(&keyReq, deviceIds)
middleAgent.AsyncEnterText(deviceIds, "Hello")
```

---

## HTTP 代理接口

ApiAgent 在 `127.0.0.2:8888` 启动 HTTP 服务，支持以下请求格式：

### 请求格式

所有请求均为 **POST**，Body 为 JSON：

```json
{
    "app": "模块名",
    "fun": "函数名",
    "data": { ... }
}
```

### 拦截的接口

以下请求会被 ApiAgent 本地处理，不转发到集控平台：

| app | fun | 说明 |
|-----|-----|------|
| `loginCtl` | `secretKeyLogin` | API Key 登录 |
| `userLogin` | `apiLogin` | API 登录（同上） |
| `userDeviceCtl` | `getUserDeviceList` | 获取设备列表 |

**登录请求示例：**

```json
POST http://127.0.0.2:8888/
{
    "app": "userLogin",
    "fun": "apiLogin",
    "data": {
        "secretKey": "你的API_KEY"
    }
}
```

**响应：**

```json
{
    "code": 0,
    "msg": "login success",
    "data": "token字符串"
}
```

### 其他请求

未被拦截的请求会自动转发到集控平台（`https://minio.accjs.cn`），并自动附加认证 Cookie。

### WebSocket 接口

```
ws://127.0.0.2:8888/?middleWareId=中间件ID
```

建立 WebSocket 连接后，ApiAgent 会自动创建对应中间件的 RTC 连接，实现 **WS ↔ RTC** 双向消息桥接。

---

## 同步 API 参考

通过中间件 RTC 连接，以同步方式（请求-响应）操作设备：

### 设备信息

| 方法 | 参数 | 说明 |
|------|------|------|
| `SyncGetAllList()` | 无 | 获取所有设备列表 |
| `SyncGetOnlineList()` | 无 | 获取在线设备列表 |
| `SyncGetDeviceDetails(deviceId)` | 设备ID | 获取设备详细信息 |
| `SyncGetPluginList()` | 无 | 获取插件列表 |
| `SyncGetMiddlewareWorkMode()` | 无 | 获取中间件工作模式 |

### 设备控制

| 方法 | 参数 | 说明 |
|------|------|------|
| `SyncModeSwitch(deviceId, mode)` | mode: 0=OTG, 1=USB | 切换设备模式 |
| `SyncDevicePowerControl(deviceId, power)` | power: 0=断电, 1=供电, 2=重启 | 电源控制（仅新机型） |
| `SyncEnterFlashingMode(deviceId, mode)` | mode: 1/2/3 | 进入刷机模式（仅新机型） |
| `SyncSetDeviceToFindMode(deviceId, mode)` | mode: 0=关, 1=开 | 设备查找模式（仅新机型） |

### 屏幕控制

| 方法 | 参数 | 说明 |
|------|------|------|
| `SyncScreenOn(deviceId)` | 设备ID | 开启屏幕 |
| `SyncScreenOff(deviceId)` | 设备ID | 关闭屏幕 |
| `SyncSwitchFront(deviceId)` | 设备ID | 切换前置摄像头 |
| `SyncSwitchBack(deviceId)` | 设备ID | 切换后置摄像头 |
| `SyncGetDeviceImage(deviceId, width, type)` | type: 0=jpg, 2=webp | 获取屏幕截图（字节） |
| `SyncGetDeviceImageToBase64(deviceId, width, type)` | 同上 | 获取屏幕截图（Base64） |

### 应用管理

| 方法 | 参数 | 说明 |
|------|------|------|
| `SyncGetAppList(deviceId)` | 设备ID | 获取已安装应用列表 |
| `SyncRunApp(deviceId, packageName)` | 包名 | 启动应用 |
| `SyncDownloadAndInstall(deviceId, install)` | DownloadAndInstall 结构体 | 下载并安装应用 |
| `SyncCheckProgress(deviceId, taskId)` | 任务ID | 查询安装进度 |
| `SyncRootApp(deviceId, pkg)` | 包名 | 授予 Root 权限 |
| `SyncCancelRootApp(deviceId, pkg)` | 包名 | 取消 Root 权限 |

### 其他

| 方法 | 参数 | 说明 |
|------|------|------|
| `SyncShellCommand(deviceId, shell)` | Shell 命令 | 执行 Shell 命令并返回结果 |
| `SyncGetClipboard(deviceId)` | 设备ID | 获取剪贴板内容 |
| `SyncDelayedetection()` | 无 | 测试延迟（毫秒） |

---

## 异步 API 参考

异步方式支持批量操作，通过回调获取结果：

| 方法 | 参数 | 说明 |
|------|------|------|
| `AsyncGetAllList()` | 无 | 获取所有设备列表 |
| `AsyncGetOnlineList()` | 无 | 获取在线列表 |
| `AsyncGetAllDeviceDetails(deviceIds)` | 设备ID数组 | 批量获取设备详情 |
| `AsyncSetAllDevicesScreenOn(deviceIds)` | 设备ID数组 | 批量开启屏幕常亮 |
| `AsyncSetAllDevicesScreenOff(deviceIds)` | 设备ID数组 | 批量关闭屏幕 |
| `AsyncTouch(touchReq, deviceIds)` | TouchRequest + 设备ID数组 | 批量触摸操作 |
| `AsyncScroll(scrollReq, deviceIds)` | ScrollRequest + 设备ID数组 | 批量滚动操作 |
| `AsyncKeyBoard(keyReq, deviceIds)` | KeyBoardRequest + 设备ID数组 | 批量键盘输入 |
| `AsyncRunAppFromPackageName(runApp, deviceIds)` | RunApp + 设备ID数组 | 批量启动应用 |
| `AsyncEnterText(deviceIds, text)` | 设备ID数组 + 文本 | 批量输入文本 |

> **注意：** `deviceIds` 为空时，默认发送给本中间件下所有设备。

---

## 端口映射

将远程设备的端口映射到本地：

```go
portMap, err := core.CreatPortMapSocket5Rtc(
    deviceId,   // 设备 ID
    9999,       // 手机端口
    9999,       // 本地端口
    0,          // 模式：0=端口映射, 1=无复用端口映射, 2=SOCKS5 代理
)
```

映射成功后，访问本地 `127.0.0.1:9999` 即等同于访问远程设备的 `9999` 端口。

---

## H264 视频串流

创建设备直连 RTC，接收 H264 视频流和音频流：

```go
device, err := core.CreatDeviceH264AudioRtc(
    deviceId,
    onOpen,     // 连接成功回调
    onClose,    // 连接断开回调
    onMessage,  // 数据回调（H264/Audio 帧）
)

// 开始视频流
device.AsyncStartVideo(reconnect)  // reconnect: 0=自动重连, 1=手动

// 开始音频流
device.AsyncStartAudio(reconnect)

// 停止
device.AsyncStopVideo()
device.AsyncStopAudio()
```

---

## DeviceId 编码规则

DeviceId 由**中间件 ID** 和**盘位号（Seat）** 组合而成：

```
DeviceId = (MiddlewareId << 8) + Seat
```

工具函数：

```go
// 从 DeviceId 提取中间件 ID
middleId := publicStruct.GetMiddleIdFromDeviceId(deviceId)

// 从 DeviceId 提取盘位号
seat := publicStruct.GetSeatFromDeviceId(deviceId)

// 从中间件 ID + 盘位号 生成 DeviceId
deviceId := publicStruct.GetDeviceIdFromMiddleIdAndSeat(middleId, seat)
```

---

## 连接状态码

### 中间件 RTC

| Code | 含义 |
|------|------|
| 1 | 连接建立 |
| 2 | 连接成功 |
| -1 | 断开连接 |
| -2 | 连接失败 |

### 端口映射

端口映射和 SOCKS5 各有独立的状态码序列，可通过 `GetCodeMsg()` 获取 JSON 格式的状态信息。

---

## 依赖

| 依赖 | 用途 |
|------|------|
| `cnb.cool/accbot/goTool` | 会话管理、WebSocket 封装、错误处理 |
| `cnb.cool/htsystem/adminApi` | 集控平台 API 客户端 |
| `github.com/ghp3000/netclient` | 网络客户端（RTC/TCP/WS/Buffer） |
| `github.com/ghp3000/public` | 消息协议、公共常量 |
| `github.com/gin-gonic/gin` | HTTP 框架 |
| `github.com/gorilla/websocket` | WebSocket 支持 |
| `github.com/pion/webrtc` | WebRTC 实现（间接依赖） |

---

## 注意事项

1. HTTP 代理监听地址为 `127.0.0.2:8888`（非 `127.0.0.1`），确保本地回环地址可用
2. 所有 RTC 连接默认自动重连（`Reconnect=0`），设置为 `1` 则断开后不重连
3. 同步调用默认超时 5 秒
4. 异步批量操作每 20 台设备分一批发送
5. 新机型专属功能（电源控制、刷机模式、查找模式）对老机箱不可用
