#!/bin/bash

# 创建目录
mkdir -p /app/status-x

# 下载程序
curl -L -o /app/status-x/status-x-client "https://newsresource.obs.cn-north-1.myhuaweicloud.com/app/download/status-x/status-x-client"

# 给予执行权限
chmod 777 /app/status-x/status-x-client

# 创建 systemd 服务文件
cat <<EOL | sudo tee /etc/systemd/system/status-x-client.service
[Unit]
Description=Status-X Service
After=network.target

[Service]
ExecStart=/usr/bin/nohup /app/status-x/status-x-client
Environment="INFLUX_TOKEN=wh56EgkTNCyt-oSz_4Uo8l_SYy9R57CnUFy2NZY4bxmjZ9bbBNiMvQ0kdo8W4cwdvP6JrgXY49uXpTI7d5mRtA=="
Environment="INFLUX_URL=10.10.18.116:8086"
Environment="SERVER_URL=10.15.97.66:42800"
Environment="COLLECT_FREQUENCY=2"
Environment="MONITOR_DISK_PATH=all"
StandardOutput=file:/app/status-x/status-x-client_info.log
StandardError=inherit
Restart=always
User=root
Group=root
WorkingDirectory=/app/status-x

[Install]
WantedBy=multi-user.target
EOL

# 使配置生效并启动服务
sudo systemctl daemon-reload
sudo systemctl enable status-x-client
sudo systemctl start status-x-client

# 检查服务状态
if systemctl is-active --quiet status-x-client; then
    echo "Status-X Client 安装并启动成功."
else
    echo "启动 Status-X Client 失败.可以查看如下日志"
    # 可选：输出服务状态信息以帮助调试
    sudo systemctl status status-x-client
fi

