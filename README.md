# filestore_server
基于Ceph的云存储项目，服务器使用gin框架搭建，使用RabbitMQ实现文件的异步转存，数据库为MySQL，前端界面使用HTML+CSS+JQuery。

Ceph集群使用docker单机搭建，参考：https://www.cnblogs.com/aganippe/p/16095588.html  
golang ceph客户端操作，参考：https://www.cnblogs.com/aganippe/p/16099067.html

* 使用Gin框架开发服务器，支持用户注册/登录、用户好友添加、文件上传下载、文件共享等功能。 
* 使用RabbitMQ实现用户上传文件异步转移到Ceph集群。 
* 使用MySQL数据库将文件信息、用户信息、用户文件对应信息、共享信息和好友请求信息进行保存。 
* 使用前端加密库CryptoJS和国密算法库gmjs，实现在前端对文件进行加解密。 
* 使用加密文件共享的传统方式，实现了多用户共享文件的功能。 
