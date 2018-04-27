#二次开发gatadvisor
##优化
    1. 舍弃run封装启动方式。
    2. 容器endpoint取自docker的hostname，而非传入变量方式。
    3. 增加hostname过滤机制。
    4. 日志采用glog，而非写日志。
    5. 自动上报采集程序的存活监控状态

##bugfix
    1. 修复index out of range。
    2. 修复因为检测自身存活状态而导致程序退出。



##采集的指标
--------------------------
| Counters | Notes|
|-----|------|
|cpu.busy|cpu使用情况百分比|
|cpu.user|用户态使用的CPU百分比|
|cpu.system|内核态使用的CPU百分比|
|cpu.core.busy|每个cpu的使用情况|
|mem.memused.percent|内存使用百分比|
|mem.memused|内存使用原值|
|mem.memtotal|内存总量|
|mem.memused.hot|内存热使用情况|
|disk.io.read_bytes|磁盘io读字节数|
|disk.io.write_bytes|磁盘io写字节数|
|net.if.in.bytes|网络io流入字节数|
|net.if.in.packets|网络io流入包数|
|net.if.in.errors|网络io流入出错数|
|net.if.in.dropped|网络io流入丢弃数|
|net.if.out.bytes|网络io流出字节数|
|net.if.out.packets|网络io流出包数|
|net.if.out.errors|网络io流出出错数|
|net.if.out.dropped|网络io流出丢弃数|

Contributors
-----------------------------------------
