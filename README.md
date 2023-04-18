<p style="text-align: center">
  <img alt="logo" height="48" src="https://openscrm.oss-cn-hangzhou.aliyuncs.com/public/openscrm_logo.svg">
</p>

<h3 style="text-align: center">
安全，强大，易开发的企业微信SCRM
</h3>

[安装](#如何安装) |

### 项目简介

> 此项目为OpenSCRM **会话存档服务** 项目

### 如何安装
### 重要提示！！！
由于依赖腾讯官方.so文件，仅支持Linux下编译，windows下可使用wsl2
#### 设置环境变量
```bash
export LD_LIBRARY_PATH=$(pwd)/lib
export GOPROXY=https://proxy.golang.com.cn,direct
```
- 复制粘贴api-server的配置
- 编译
```bash
CGO_ENABLED=1 go build -o msg-arch-server main.go
```

### 联系作者

<img src="https://openscrm.oss-cn-hangzhou.aliyuncs.com/public/screenshots/qrcode.png" width="200" />

扫码可加入交流群

### 版权声明

OpenSCRM遵循Apache2.0协议，可免费商用
