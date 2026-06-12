# syntax=docker/dockerfile:1

# ── 阶段 1: 前端构建 ──────────────────────────────────────────────────────────
FROM node:20-slim AS frontend
WORKDIR /web
COPY web/package*.json ./
RUN npm ci --prefer-offline
COPY web/ ./
# Vite 产物直接输出到 dist/，再由下一阶段 COPY 进 Go embed 目录
RUN npm run build

# ── 阶段 2: Go 后端编译 ───────────────────────────────────────────────────────
# gopsutil/v4 需 Go ≥ 1.24；CGO 需 gcc（golang 官方镜像自带）
FROM golang:1.25-bookworm AS builder
WORKDIR /src
ENV CGO_ENABLED=1 \
    GOFLAGS=-mod=vendor
# 依赖已 vendor，无需网络访问
COPY go.mod go.sum vendor/ ./
COPY . .
# 用 Vite 产物替换占位 dist/
COPY --from=frontend /web/dist ./internal/server/dist
# go-nvml 编译期不需要 CUDA/驱动(nvml.h 已 vendored)，运行时 dlopen libnvidia-ml.so.1
RUN go build -trimpath -ldflags="-s -w" -o /out/gpu-monitor ./cmd/monitor

# ── 阶段 3: 运行时 ────────────────────────────────────────────────────────────
# CGO 动态二进制 → 必须 glibc 基础镜像（不能用 alpine/scratch）
FROM debian:bookworm-slim
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates tzdata \
 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /out/gpu-monitor /usr/local/bin/gpu-monitor
EXPOSE 8800
ENTRYPOINT ["/usr/local/bin/gpu-monitor"]
CMD ["-config", "/etc/gpu-monitor/config.yaml"]
