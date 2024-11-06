

docker run -d --name=influxdb -p 8086:8086 influxdb:2


docker build -t status-x:0.0.2 .
docker tag status-x:0.0.2 10.15.98.150/library-hf/status-x:0.0.2
docker push 10.15.98.150/library-hf/status-x:0.0.1

docker build -t status-x:0.0.4 .
docker tag status-x:0.0.4 dockerhubbs.finchina.com/finchina-dev/status-x:0.0.4
docker push dockerhubbs.finchina.com/finchina-dev/status-x:0.0.4


docker run -d --name mysql5.8 -p 13306:3306 -e MYSQL_ROOT_PASSWORD=11111111 -v /app/status-x/mysqlDate:/var/lib/mysql --restart unless-stopped mysql:5.8 --default-authentication-plugin=mysql_native_password --bind-address=0.0.0.0

docker login -u tcdept-hf -p Tcdept@427 dockerhubbs.finchina.com
116 服务器

docker build -t status-x-server:0.0.5 .
docker tag status-x-server:0.0.5 dockerhubbs.finchina.com/finchina-dev/status-x-server:0.0.5
docker push dockerhubbs.finchina.com/finchina-dev/status-x-server:0.0.5



指标解释
jvm.memory.max：JVM 可使用的最大内存量。了解应用程序的内存上限非常有用。

jvm.memory.used：JVM 当前使用的内存量。监控此项有助于检测内存泄漏并确保内存使用效率。

jvm.memory.committed：JVM 从操作系统获取并保证可用的内存量，即已分配给 JVM 的内存。

jvm.gc.pause：垃圾回收暂停所花费的时间。较高的值可能表示由于频繁或长时间的垃圾回收导致的性能问题。

jvm.gc.memory.promoted：在垃圾回收期间，从新生代移动到老年代的内存量。

jvm.gc.max.data.size：老年代内存池的最大大小。

jvm.gc.live.data.size：完整垃圾回收后老年代内存池的大小。

jvm.gc.memory.allocated：为垃圾回收目的分配的内存量。

jvm.classes.loaded：当前加载到 JVM 中的类的数量。

jvm.classes.unloaded：自 JVM 启动以来已卸载的类的总数。

jvm.threads.live：当前活动线程的数量。

jvm.threads.daemon：守护线程的数量。

jvm.threads.peak：自 JVM 启动以来或自峰值重置以来的线程峰值数。

jvm.threads.states：所有线程的当前状态（例如：RUNNABLE、BLOCKED、WAITING）。

jvm.buffer.count：直接和映射缓冲区的数量。

jvm.buffer.memory.used：直接和映射缓冲区使用的内存。

jvm.buffer.total.capacity：所有直接和映射缓冲区的总容量。

system.cpu.count：JVM 可用的处理器数量。

system.cpu.usage：整个系统的“最近 CPU 使用率”。

system.load.average.1m：最近一分钟的系统平均负载。

process.cpu.usage：Java 虚拟机进程的“最近 CPU 使用率”。

process.uptime：Java 虚拟机的运行时间。

process.start.time：进程的启动时间。

process.files.max：进程可以打开的最大文件描述符数量。

process.files.open：当前打开的文件描述符数量。

http.server.requests：与 HTTP 请求相关的指标，包括计数、响应时间和状态码。

logback.events：Logback 日志事件的数量。用于监控日志活动。

tomcat.sessions.created：Tomcat 服务器创建的会话总数。

tomcat.sessions.expired：已过期的会话总数。

tomcat.sessions.active.current：当前活跃的会话数量。

tomcat.sessions.active.max：服务器启动以来的最大活跃会话数量。

tomcat.sessions.alive.max：会话存活的最长时间（以毫秒为单位）。

tomcat.sessions.rejected：由于达到会话限制而被拒绝的会话数量。