一个特别简单的cronweb的demo

项目结构:  
crontab  
    /master  
    /worker  
    /command  
master  
    1.搭建go项目框架，配置文件，命令行参数，线程配置  
    2.给web后台提供HTTP API，用于管理job  
    3.写一个web后台的前端页面，bootstrap+jq，前后端分离开发  
 
 worker  
    1.从etcd中把job同步到内存中  
    2.实现调度模块，基于cron表达式调度N个job  
    3.实现执行模块，并发的执行多个job  
    4.对job的分布式锁，防止集群并发  
    5.把执行日志保存到mongoDB  
    