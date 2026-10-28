#!/usr/bin/env bash
# Blog 后端启动脚本（流水线会把它安装到 /opt/blog/backend/start.sh 并 chmod +x）
# 用法：/opt/blog/backend/start.sh
# 先杀掉旧进程，再加载 backend.env，nohup 后台启动
set -e
cd "$(dirname "$0")"
# 用 -x 精确匹配进程名：-f 会匹配任何命令行含 "blog-backend" 的进程，
# 包括执行部署的 ssh 会话本身，导致部署脚本被中途杀掉误报失败
pkill -x blog-backend || true
sleep 1
[ -f backend.env ] && set -a && source backend.env && set +a
nohup ./blog-backend > blog.log 2>&1 &
