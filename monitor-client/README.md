

set GOARCH=amd64
go env -w GOARCH=amd64
set GOOS=linux
go env -w GOOS=linux

go build -o status-x-client
ps aux | grep status-x-client


sudo vi /etc/systemd/system/status-x-client.service

[Unit]
Description=Status-X Service
After=network.target

[Service]
ExecStart=/usr/bin/nohup /app/status-x/status-x-client
Environment="INFLUX_TOKEN=wh56EgkTNCyt-oSz_4Uo8l_SYy9R57CnUFy2NZY4bxmjZ9bbBNiMvQ0kdo8W4cwdvP6JrgXY49uXpTI7d5mRtA=="
Environment="INFLUX_URL=10.10.18.116:8086"
Environment="SERVER_URL=10.10.18.116:42800"
Environment="COLLECT_FREQUENCY=2"
Environment="MONITOR_DISK_PATH=/app"
StandardOutput=file:/app/status-x/status-x-client_info.log
StandardError=inherit
Restart=always
User=root
Group=root
WorkingDirectory=/app/status-x

[Install]
WantedBy=multi-user.target

sudo systemctl daemon-reload
sudo systemctl enable status-x-client
sudo systemctl start status-x-client
sudo systemctl stop status-x-client
sudo systemctl disable status-x-client
sudo systemctl status status-x-client




INFLUX_TOKEN="wh56EgkTNCyt-oSz_4Uo8l_SYy9R57CnUFy2NZY4bxmjZ9bbBNiMvQ0kdo8W4cwdvP6JrgXY49uXpTI7d5mRtA==" INFLUX_URL="10.10.18.116:8086" SERVER_URL="10.10.18.116:42800" COLLECT_FREQUENCY=2 ./status-x-client


bash <(curl -s https://newsresource.obs.cn-north-1.myhuaweicloud.com/app/download/status-x/reinstall_status_x_client.sh)
bash <(curl -s https://newsresource.obs.cn-north-1.myhuaweicloud.com/app/download/status-x/install_status_x_client.sh)
