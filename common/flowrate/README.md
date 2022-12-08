# 介绍

参考 <https://github.com/mxk/go-flowrate> 项目设计的简单流量控制模块，
不同于前人的设计思路，我们严格执行了流量上限规则，即一旦达到传输上限，那么就会等待，保证数据传输的平均速率
不高于设置的流量上限值。

流量控制模块可以防止自己发送数据据过快，对方来不及处理，或者对方发送数据过快，自己来不及处理的情况发生。这在区块链网络中还是很有必要的。

```
{"Start":"2022-12-08T15:50:51.26+08:00","Bytes":25600000,"Samples":5,"CurRate":5235199,"AvgRate":5120000,"PeakRate":7422694,"Duration":5000000000,"Active":true}
{"Start":"2022-12-08T15:50:51.26+08:00","Bytes":46080000,"Samples":9,"CurRate":6499108,"AvgRate":5120000,"PeakRate":8296452,"Duration":9000000000,"Active":true}
{"Start":"2022-12-08T15:50:51.26+08:00","Bytes":71680000,"Samples":14,"CurRate":5163192,"AvgRate":5120000,"PeakRate":8296452,"Duration":14000000000,"Active":true}
{"Start":"2022-12-08T15:50:51.26+08:00","Bytes":97280000,"Samples":19,"CurRate":5970339,"AvgRate":5120000,"PeakRate":8296452,"Duration":19000000000,"Active":true}
{"Start":"2022-12-08T15:50:51.26+08:00","Bytes":122880000,"Samples":24,"CurRate":6225263,"AvgRate":5120000,"PeakRate":8296452,"Duration":24000000000,"Active":true}
{"Start":"2022-12-08T15:50:51.26+08:00","Bytes":148480000,"Samples":29,"CurRate":6952823,"AvgRate":5120000,"PeakRate":8516404,"Duration":29000000000,"Active":true}
{"Start":"2022-12-08T15:50:51.26+08:00","Bytes":174080000,"Samples":34,"CurRate":7873000,"AvgRate":5120000,"PeakRate":8516404,"Duration":34000000000,"Active":true}
```