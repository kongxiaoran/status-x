

docker run -d --name=influxdb -p 8086:8086 influxdb:2


docker build -t status-x:0.0.2 .
docker tag status-x:0.0.2 10.15.98.150/library-hf/status-x:0.0.2
docker push 10.15.98.150/library-hf/status-x:0.0.1

docker build -t status-x:0.0.2 .
docker tag status-x:0.0.2 dockerhubbs.finchina.com:443/finchina-dev/status-x:0.0.2
docker push dockerhubbs.finchina.com:443/finchina-dev/status-x:0.0.2 


docker run -d --name mysql5.8 -p 13306:3306 -e MYSQL_ROOT_PASSWORD=11111111 -v /app/status-x/mysqlDate:/var/lib/mysql --restart unless-stopped mysql:5.8 --default-authentication-plugin=mysql_native_password --bind-address=0.0.0.0


docker build -t status-x-server:0.0.2 .
docker tag status-x-server:0.0.2 dockerhubbs.finchina.com:443/finchina-dev/status-x-server:0.0.2
docker push dockerhubbs.finchina.com:443/finchina-dev/status-x-server:0.0.2 