# CLAUDE.md

本文件为 Claude Code 在本仓库工作的基线约定。所有架构决策、工程约束、构建方式均已**经真机/编译验证**(RTX 5090 + 驱动 595.71.05 + NVML 13.x + go-nvml v0.13.1-0 + gopsutil v4),非推测,改动前请遵循。

---

## 1. 项目概述

`gpu-monitor` 是一个 **Go 单二进制系统监控服务**,在浏览器里提供接近 **btop + nvtop** 的实时体验,并把**阈值告警 + 定时汇总**推送到 Telegram。

- Module: `github.com/windskyer/gpu-monitor`
- 目标主机:**单机部署**,Ubuntu,本机 RTX 5090 (Blackwell GB202) + Ryzen 9 9950X
- **部署形态:Docker 容器(容器内编译)**,经 NVIDIA Container Toolkit 访问 GPU
- 不做多机/agent 聚合(采集层用接口抽象预留,但当前不实现)

---

## 2. 技术栈与锁定决策(已验证,勿擅改)

| 项 | 决策 | 验证结论 |
|---|---|---|
| GPU 数据源 | `github.com/NVIDIA/go-nvml` **v0.13.1-0** | 已编译通过,API 签名匹配 |
| 系统指标 | `github.com/shirou/gopsutil/v4`(需 **Go ≥ 1.24**) | v4.26.5 实测要求 1.24 |
| 编译 | **`CGO_ENABLED=1` 必须**(0 会报 `undefined: Return`) | 编译期**不需要 CUDA/驱动**,`nvml.h` 已 vendored |
| 运行时库 | **dlopen `libnvidia-ml.so.1`**,二进制不链接 nvidia 库 | 由 Container Toolkit 注入 |
| 基础镜像 | builder `golang:1.24-bookworm`;runtime **`debian:bookworm-slim`(glibc)** | CGO 动态二进制**不能用 alpine/scratch** |
| 实时通道 | **WebSocket**,服务端 1Hz 推 `Snapshot` JSON | |
| 前端 | **vanilla JS + `<canvas>` + `go:embed`**,零 npm 构建 | |
| 配置 | YAML | |
| 容器权限 | `--gpus all` + `pid: host`,宿主 `/proc`·`/sys` 只读挂载 | 进程级 GPU 需 host PID 命名空间 |
| Prometheus | 内置,默认关闭 | M3 |

---

## 3. NVML 字段约束(基于 5090 真机实测)

实测全部 `OK`,**唯一例外**:`FI_DEV_MEMORY_TEMP`(显存结温)返回 `Not Supported` —— 消费级卡不暴露,**不要采集此字段**。

限频原因:`GetCurrentClocksThrottleReasons` 在 NVML 13 已 **DEPRECATED**。
**统一使用 `GetCurrentClocksEventReasons`**,Throttle 版本仅作老驱动兜底(当前不需要)。

锁定字段集:

```
标识   : GetName / GetUUID
负载   : GetUtilizationRates (GPU% + Mem%)
显存   : GetMemoryInfo (used/total/free)
温度   : GetTemperature(TEMPERATURE_GPU)        # 仅核心温度,无结温
功耗   : GetPowerUsage / GetEnforcedPowerLimit  # 单位 mW,需 /1000
风扇   : GetFanSpeed
时钟   : GetClockInfo(GRAPHICS/SM/MEM)
编解码 : GetEncoderUtilization / GetDecoderUtilization
PCIe   : GetCurrPcieLinkGeneration/Width + GetPcieThroughput(TX/RX_BYTES)
限频   : GetCurrentClocksEventReasons           # 位掩码,需解码
进程   : GetComputeRunningProcesses + GetGraphicsRunningProcesses  # M2,需 host PID
```

---

## 4. 容器化与宿主指标采集(关键)

容器内默认只能看到容器自己的命名空间。要做 btop 那样的**宿主监控**,必须按下表处理。**GPU 指标天然是宿主级**(NVML 直连驱动),无需特殊处理;系统指标才需要。

| 指标 | 机制 | 容器要求 |
|---|---|---|
| CPU/内存/负载 | gopsutil 读 `HOST_PROC` | 挂 `/proc:/host/proc:ro` + `HOST_PROC=/host/proc` |
| CPU 温度(k10temp) | gopsutil 读 `HOST_SYS`/hwmon | 挂 `/sys:/host/sys:ro` + `HOST_SYS=/host/sys` |
| 网络带宽 | gopsutil 读 `HOST_PROC/net/dev` | 同 HOST_PROC,**无需** `--net host` |
| 磁盘分区 | gopsutil 读 `HOST_PROC/1/mountinfo` | 同 HOST_PROC |
| 磁盘用量 | **gopsutil 无挂载前缀变量** | 挂 `/:/host/root:ro,rslave`,**本项目代码自行加前缀**(见下) |
| 进程列表 / GPU 进程名 | NVML 返回**宿主 PID**,需读宿主 /proc | `pid: host` |

**磁盘用量前缀(必须在代码里实现)**:gopsutil 没有 `HOST_MOUNT_PREFIX`。disk 采集器读到宿主挂载点(如 `/`、`/data`)后,调用 `disk.Usage()` 时要前缀配置项 `host.mount_prefix`(容器内为 `/host/root`,裸机为空)。即 `disk.Usage(cfg.Host.MountPrefix + mountpoint)`。`/` 挂载用 `ro,rslave` 以便宿主子挂载(独立盘)可见。

`HOST_PROC`/`HOST_SYS`/`HOST_ETC` 由 gopsutil **自己读环境变量**,我们只需在容器里设好;`host.mount_prefix` 是**本项目自有配置**,只作用于磁盘用量。

---

## 5. 目录结构

```
cmd/monitor/main.go     入口:加载配置、初始化 NVML、启动 scheduler 与 server
internal/config/        YAML 加载与校验(含 host.mount_prefix)
internal/model/         Snapshot 及时序类型(纯数据,无外部依赖)
internal/collector/     Collector 接口 + cpu/mem/disk/net/proc 实现(gopsutil)
internal/gpu/           NVML 封装(go-nvml),实现 Collector 接口
internal/store/         ring-buffer 历史 + 上周期计数器(算速率)
internal/server/        HTTP + WebSocket + go:embed(web/)
internal/alert/         规则引擎 + 状态机(idle/firing/resolved)
internal/notify/        Telegram 客户端
internal/exporter/      Prometheus(M3,默认关)
web/                    内嵌静态资源(index.html / app.js / style.css)
Dockerfile              多阶段:golang:1.24 编译 → debian-slim 运行
docker-compose.yml      GPU + 宿主挂载 + 端口映射
```

依赖方向:`model` 被所有人依赖,自身零依赖;`collector`/`gpu` 只产出 `model`;`server`/`alert` 只消费 `model`。禁止反向依赖。

---

## 6. 核心数据流

```
ticker(1s) → Scheduler 采集 → Snapshot → 扇出到:
  ├─ Store        (ring-buffer + 速率计算)
  ├─ WS broadcaster (推所有连接)
  ├─ Alerter      (规则评估 → Notify)
  └─ Exporter     (Prometheus,可选)
Reporter 独立 ticker → 定时 TG 汇总
进程列表独立降频采样(2s),结果缓存供主快照引用
```

**速率类指标**(net/disk IO/PCIe throughput)由 Scheduler 持有上周期计数器,用 `(cur-prev)/dt` 计算,不要在采集器里自行算。

---

## 7. 构建 / 运行(容器)

```bash
# 构建镜像(容器内编译,无需宿主装 Go)
docker compose build          # 或 docker build -t gpu-monitor:latest .

# 运行
docker compose up -d

# 前提:宿主已装 NVIDIA Container Toolkit 且 Docker 配好 nvidia runtime
#   nvidia-ctk runtime configure --runtime=docker && systemctl restart docker
```

go-nvml 锁版本:`go get github.com/NVIDIA/go-nvml/pkg/nvml@v0.13.1-0`
国内拉依赖:Dockerfile 内已设 `GOPROXY=https://goproxy.cn,direct`。

**容器内 listen 必须为 `0.0.0.0:8800`**(容器网络),对外通过 `ports: 127.0.0.1:8800:8800` 限制为仅本机,再经宿主 nginx 反代 + token 暴露。**不要**用 `--net host` 后再绑 0.0.0.0 裸暴露。

---

## 8. 配置(config.yaml)

```yaml
server:   { listen: "0.0.0.0:8800", token: "change-me" }   # 容器内绑 0.0.0.0,靠端口映射限制
host:     { mount_prefix: "/host/root" }                    # 容器=/host/root;裸机=""
sample:   { system: 1s, gpu: 1s, process: 2s, history: 180, top_n: 15 }
telegram: { bot_token: "", chat_id: "", report_interval: 1h }
alerts:
  gpu_temp:    { threshold: 83, cooldown: 10m }
  gpu_mem_pct: { threshold: 95, cooldown: 10m }
  cpu_temp:    { threshold: 85, cooldown: 10m }
  mem_pct:     { threshold: 90, cooldown: 10m }
  disk_pct:    { threshold: 90, cooldown: 30m }
exporter: { prometheus: false }
```

CPU 温度:9950X 走 `k10temp`(Tctl/Tdie),sensorKey 命名按内核版本兜底匹配。

---

## 9. 编码约定

- Go 1.24+,`gofmt`/`go vet` 干净;错误用 `fmt.Errorf("...: %w", err)` 包装。
- 采集器对**单字段失败要优雅降级**(该字段置零/nil 并记 debug 日志),**不得 panic**,不得因一个 GPU 字段失败而中断整次采集。
- **磁盘用量必须经 `host.mount_prefix` 前缀**;裸机模式前缀为空,容器模式为 `/host/root`。
- 非 root/无相应挂载时,受限字段返回错误是**预期行为**,隐藏对应数据,不报错退出。
- NVML 调用统一检查 `Return != nvml.SUCCESS`;`Init`/`Shutdown` 成对,`Shutdown` 用 `defer`。
- 并发:Snapshot 一旦产出即不可变(只读扇出);WS 写带超时,卡住的连接直接踢。
- 前端零第三方框架/CDN,所有资源 `go:embed` 进二进制。

---

## 10. 里程碑

- **M1(当前目标)**:CPU/内存/磁盘/网络基础 + GPU 基础字段(第 3 节,除进程外)+ Scheduler/Store + WS 面板 + 阈值告警 + 定时汇总 + YAML + 容器化(Dockerfile/compose)。
- **M2**:每核 CPU、系统进程表、**GPU 进程级**、磁盘 IO/网络曲线、enc/dec/PCIe/限频解码、恢复通知。
- **M3**:Prometheus `/metrics`、`/api/history`、token 鉴权完善、配置热加载。

---

## 11. Do / Don't

- ✅ 改 NVML 相关代码前,以第 3 节实测字段集为准。
- ✅ 容器内采集宿主指标,严格按第 4 节挂载与 `HOST_*` 处理。
- ✅ 磁盘用量加 `host.mount_prefix` 前缀。
- ✅ 速率计算放 Store/Scheduler,采集器只取瞬时值。
- ✅ 保持单二进制、零前端构建。
- ❌ 不要采集显存结温(`FI_DEV_MEMORY_TEMP`,本卡不支持)。
- ❌ 不要用 `GetCurrentClocksThrottleReasons`(已废弃),用 Event 版本。
- ❌ 不要设 `CGO_ENABLED=0`(编译失败)。
- ❌ runtime 镜像**不要用 alpine/scratch**(CGO 动态二进制需 glibc)。
- ❌ 不要把驱动/CUDA 打进镜像(由 Container Toolkit 运行时注入)。
- ❌ 不要引入多机/agent 架构,不要加 npm 前端构建链。
- ❌ 不要 `--net host` 后裸绑 0.0.0.0;用端口映射限本机 + nginx 反代。
