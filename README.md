# Qnet

- `master` 基于 Golang net包的网络库
- `feature-gnet` 基于潘剑锋大神的 [gnet库](https://github.com/panjf2000/gnet)
- 轻度封装
- 屏蔽底层协议，切换协议只需修改启动参数即可
- 消息注册模式，方便使用
- 支持多种协议
    - [x] tcp
    - [x] udp
    - [x] ws
    - [ ] kcp

## Tcp
- 支持 tcp 包头自定义


## TODO List
- [x] ws 支持 ticker
- [x] ws 支持返回 action