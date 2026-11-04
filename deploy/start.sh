#!/usr/bin/env bash
# Blog 后端启动脚本（流水线会把它安装到 /opt/blog/backend/start.sh 并 chmod +x）
# 由 systemd 托管（ExecStart=/opt/blog/backend/start.sh）：开机自启、崩溃自动拉起
# 不再自行 pkill / nohup，进程生命周期交给 systemd 管理
set -e
cd "$(dirname "$0")"
if [ -f backend.env ]; then
  set -a
  # shellcheck disable=SC1091
  source backend.env
  set +a
fi
# exec 让 Go 进程直接接管 PID，便于 systemd 跟踪与优雅停机
exec ./blog-backend
