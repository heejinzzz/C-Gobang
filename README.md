# C-Gobang
多人实时五子棋对战游戏

---

## 0. 直接使用
点击链接 [C_Gobang.apk 下载](https://pan.xunlei.com/s/VNCyoSPqTtkvTCYFWerEarHbA1)，输入提取码：z3nf，下载 C_Gobang.apk 文件并安装到安卓设备上，直接开始使用 C-Gobang App。

> 注：当前服务器可能会因过期而停止服务。对此，可以根据下述步骤部署自己的 C-Gobang 服务端，并编译生成与之配套的 C-Gobang 客户端。

## 1. 部署 C-Gobang 服务端（On Linux）
#### ( 1 ) 安装 docker engine
参考：[Install Docker Engine](https://docs.docker.com/engine/install/)
#### ( 2 ) 下载 docker-compose
参考：[Install Compose](https://docs.docker.com/compose/install/linux/)（推荐安装版本：v2.7.0）
#### ( 3 ) 执行命令：
```shell
git clone https://github.com/heejinzzz/C-Gobang.git
cd C-Gobang
bash deploy.sh 
```
至此，C-Gobang 服务端部署完成。

---
#### ( 附1 ) C-Gobang 服务端架构
![C-Gobang 服务端架构图](https://github.com/heejinzzz/C-Gobang/blob/main/architecture.png)
userManager、gameManager 是基于 grpc 架构的微服务。

图中每个服务实例都基于一个 docker 容器部署。

---
#### ( 附2 ) C-Gobang 服务实例相关配置
| 服务实例 | 登录用户名 | 登录密码 |
| :---: | :---: | :---: |
| mysql-master | root | 518315 |
| mysql-slave | root | 518315 |
| protainer | admin | 000000000000 |

各个服务实例的端口号可以在 docker 或者 portainer 中查看端口映射。

---
## 2. 编译生成 C-Gobang 客户端
#### ( 1 ) 确保完成安装 fyne 所需的前置条件
参考：[Prerequisites of installation fyne](https://developer.fyne.io/started/#prerequisites)
#### ( 2 ) 执行命令安装 fyne cmd 工具：
```shell
go install fyne.io/fyne/v2/cmd/fyne@latest
```
#### ( 3 ) 执行命令获取源码：
```shell
git clone https://github.com/heejinzzz/C-Gobang.git
cd C-Gobang/App
go mod tidy
```
#### ( 4 ) 修改服务端地址：
将 C-Gobang/App/config.go 中的
```go
ServerIP = "127.0.0.1"
```
中的 127.0.0.1 修改为你部署的 C-Gobang 服务端所在的服务器IP。
#### ( 5 ) 编译生成 C-Gobang 客户端
(a) 执行命令，编译生成运行在 windows 系统的 C-Gobang.exe :

    fyne package --name C-Gobang --os windows --icon icon.png

(b) 执行命令，编译生成运行在 android 系统的 C_Gobang.apk :

    fyne package --name C-Gobang --os android --appID com.heejinzzz.C_Gobang --icon icon.png
    
(c) 执行命令，编译生成运行在 ios 系统的 C_Gobang.app :

    fyne package --name C-Gobang --os ios --appID com.heejinzzz.C_Gobang --icon icon.png
