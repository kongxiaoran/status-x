

docker run -d --name=influxdb -p 8086:8086 influxdb:2


docker build -t status-x:0.0.1 .
docker tag status-x:0.0.1 10.15.98.150/library-hf/status-x:0.0.1
docker push 10.15.98.150/library-hf/status-x:0.0.1