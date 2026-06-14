# GPU Monitor

浏览器实时监控面板，提供接近 **btop + nvtop** 的体验，并通过 Telegram 推送阈值告警与定时汇总。  
单 Go 二进制，前端 Vue 3 打包内嵌，Docker 一键部署。

---

## 功能

| 模块 | 内容 |
|---|---|
| **GPU** | 利用率、显存、温度、功耗、风扇、时钟、PCIe 带宽、限频原因（RTX 5090 实测） |
| **CPU** | 使用率、温度（k10temp / coretemp）、负载、主频 |
| **内存** | RAM / Swap 用量与百分比 |
| **网络** | 各网卡实时收发速率（双线时序图） |
| **磁盘** | 各挂载点用量与 I/O 速率 |
| **告警** | 阈值触发 → Telegram 通知，支持冷却时间 |
| **汇报** | 定时（默认 1h）推送 Telegram 状态摘要 |

---

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.25，`CGO_ENABLED=1`，单二进制 |
| GPU 数据 | [go-nvml](https://github.com/NVIDIA/go-nvml) v0.13.1-0，运行时 `dlopen libnvidia-ml.so.1` |
| 系统数据 | [gopsutil](https://github.com/shirou/gopsutil) v4 |
| 实时推送 | WebSocket，服务端 1 Hz |
| 前端 | Vue 3 + Vite + TypeScript，`go:embed` 打进二进制 |
| 容器 | Node 20（前端构建）→ golang:1.25（后端编译）→ debian:bookworm-slim（运行） |

---

## 快速开始

### 前置条件

- Docker + [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html)
- 配置 nvidia runtime：

```bash
nvidia-ctk runtime configure --runtime=docker
systemctl restart docker
```

### 部署

```bash
git clone <repo-url>
cd gpu-monitor

# 复制并按需修改配置
cp config.yaml.example config.yaml   # 或直接编辑 config.yaml

# 构建并启动（三阶段 Dockerfile：Node 20 → Go 1.25 → debian-slim）
docker compose up -d --build

# 查看日志
docker compose logs -f
```

默认监听 `127.0.0.1:8800`，在同一台机器打开：

```
http://localhost:8800
```

---

## 配置

`config.yaml`（挂载到容器 `/etc/gpu-monitor/config.yaml`）：

```yaml
server:
  listen: "0.0.0.0:8800"
  domain: "https://home.flftuu.com/gpu"
  token: "change-me"         # 留空则不鉴权
  network_url: "https://home.flftuu.com/ui/"

host:
  mount_prefix: "/host/root" # 容器模式；裸机部署填 ""

sample:
  system:  1s
  gpu:     1s
  process: 2s
  history: 180               # 环形缓冲保留帧数
  top_n:   15

telegram:
  bot_token: ""
  chat_id:   ""
  report_interval: 1h        # 定时汇报间隔；留空禁用

alerts:
  gpu_temp:    { threshold: 83,  cooldown: 10m }
  gpu_mem_pct: { threshold: 95,  cooldown: 10m }
  cpu_temp:    { threshold: 85,  cooldown: 10m }
  mem_pct:     { threshold: 90,  cooldown: 10m }
  disk_pct:    { threshold: 90,  cooldown: 30m }

exporter:
  prometheus: false
```

---

## 目录结构

```
gpu-monitor/
├── cmd/monitor/          # 入口：加载配置、初始化、启动
├── internal/
│   ├── alert/            # 阈值规则引擎（idle → firing → cooldown）
│   ├── collector/        # CPU / 内存 / 磁盘 / 网络（gopsutil）
│   ├── config/           # YAML 加载与默认值
│   ├── gpu/              # NVML 封装（go-nvml）
│   ├── model/            # Snapshot 数据结构（无外部依赖）
│   ├── notify/           # Telegram 客户端 + 定时汇报
│   ├── server/           # HTTP + WebSocket + go:embed
│   │   └── dist/         # Vite 产物（npm run build 生成，git 跟踪占位）
│   └── store/            # 环形缓冲 + Scheduler（采集 → 扇出）
├── web/                  # 前端（Vue 3 + Vite + TypeScript）
│   └── src/
│       ├── composables/useWS.ts   # WebSocket + 指数退避重连
│       ├── components/
│       │   ├── TimeChart.vue      # canvas 时序图
│       │   ├── GPUPanel.vue
│       │   ├── CPUPanel.vue
│       │   ├── MemPanel.vue
│       │   ├── NetPanel.vue
│       │   └── DiskPanel.vue
│       ├── types.ts       # 与 Go model 对应的 TS 接口
│       └── utils.ts       # 格式化 / 颜色工具
├── vendor/                # Go 依赖（vendored，Docker 构建无需网络）
├── config.yaml            # 运行配置
├── docker-compose.yml
└── Dockerfile             # 三阶段构建
```

---

## 开发

### 后端

```bash
# 需要 Go 1.25+，CGO_ENABLED=1（go-nvml 依赖 CGO）
go run ./cmd/monitor -config config.yaml
```

> 裸机开发时将 `config.yaml` 中 `host.mount_prefix` 改为 `""`。

### 前端

```bash
cd web
npm install
npm run dev        # Vite dev server → http://localhost:5173
                   # /ws 和 /api 自动代理到 Go :8800
```

修改 `.vue` / `.ts` 文件后 HMR 即时热更新，无需重启。

### 生产构建（仅前端）

```bash
cd web
npm run build
# 产物输出到 ../internal/server/dist/，下次 go build 自动内嵌
```

---

## 容器挂载说明

`docker-compose.yml` 已配置完整宿主挂载，容器内可读取宿主指标：

| 挂载 | 用途 |
|---|---|
| `/proc:/host/proc:ro` + `HOST_PROC=/host/proc` | CPU / 内存 / 网络 / 进程 |
| `/sys:/host/sys:ro` + `HOST_SYS=/host/sys` | CPU 温度（hwmon） |
| `/:/host/root:ro,rslave` | 磁盘用量（`host.mount_prefix=/host/root`） |
| `pid: host` | GPU 进程名解析（NVML 返回宿主 PID） |
| `NVIDIA_VISIBLE_DEVICES: all` | NVIDIA Container Toolkit 注入 `libnvidia-ml.so.1` |

---

## 里程碑

- **M1（当前）** CPU / 内存 / 磁盘 / 网络 + GPU 基础字段 + WebSocket 面板 + 阈值告警 + 定时汇总 + 容器化
- **M2** 每核 CPU、进程表、GPU 进程级、磁盘 IO 曲线、enc/dec/PCIe/限频解码、恢复通知
- **M3** Prometheus `/metrics`、`/api/history`、token 鉴权完善、配置热加载
