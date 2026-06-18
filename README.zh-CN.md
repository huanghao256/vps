# VPS Inspector

[English](README.md) | 简体中文

VPS Inspector 是一个轻量级、自托管的 Linux VPS 控制面板和质量检测工具。它直接运行在需要检测的 VPS 上，通过 Web 界面展示实时系统状态、VPS 线路质量检测结果，以及防火墙端口控制能力。

这个项目按开源可维护标准设计：核心依赖少、模块边界清晰、系统命令调用受控，前端资源可以嵌入到单个 Go 二进制文件中，方便发布和一键安装。

## 当前功能

- Go 后端，支持优雅关闭
- Token 鉴权，默认安全绑定本地地址
- 基于 Linux `/proc` 的实时系统状态
- VPS 质量检测：线路、延迟、带宽、稳定性、IP 风控风险
- 防火墙端口控制：支持 `ufw`、`firewalld`、`nftables`、`iptables`
- React + TypeScript 模块化前端
- Dockerfile、CI、Release 自动构建和项目文档

## 平台要求

VPS Inspector 只面向 Linux VPS 环境。

Windows 不是支持的运行环境。当前项目中的部分本地构建命令可在 Windows 上执行，但系统状态、防火墙控制等核心功能依赖 Linux。

## 一键安装

发布 GitHub Release 后，用户不需要安装 Go 或 Node，可以直接执行：

```bash
curl -fsSL https://raw.githubusercontent.com/huanghao256/vps/main/scripts/install.sh | sudo sh
```

自定义监听地址和访问令牌：

```bash
VPS_INSPECTOR_ADDR=0.0.0.0:8719 VPS_INSPECTOR_AUTH_TOKEN=your-token curl -fsSL https://raw.githubusercontent.com/huanghao256/vps/main/scripts/install.sh | sudo -E sh
```

如需覆盖项目根目录：

```bash
VPS_CONTROL_PANEL_HOME=/vps-control-panel curl -fsSL https://raw.githubusercontent.com/huanghao256/vps/main/scripts/install.sh | sudo -E sh
```

如果自动识别公网 IP 不准确，可以显式指定：

```bash
VPS_INSPECTOR_PUBLIC_HOST=8.163.47.95 curl -fsSL https://raw.githubusercontent.com/huanghao256/vps/main/scripts/install.sh | sudo -E sh
```

安装脚本会自动完成：

- 下载最新 Release 中的 Linux 二进制文件
- 创建 `/vps-control-panel` 项目根目录
- 安装到 `/vps-control-panel/bin/vps-inspector`
- 写入 `/vps-control-panel/config/vps-inspector.env`
- 创建 systemd 服务
- 设置开机自启
- 启动 `vps-inspector`
- 输出带令牌的访问链接

安装完成后会输出类似：

```text
Access URL: http://<服务器IP>:8719/<自动生成的令牌>
```

直接打开这个链接即可。前端会自动保存令牌，界面里不再提供手动修改访问令牌的输入框。

项目自有文件会统一放在：

```text
/vps-control-panel/bin/       二进制文件
/vps-control-panel/config/    环境配置
/vps-control-panel/systemd/   systemd 服务文件源
/vps-control-panel/data/      预留运行数据
/vps-control-panel/logs/      预留日志目录
/vps-control-panel/tmp/       运行临时文件
```

唯一位于目录外的是 systemd 入口链接：

```text
/etc/systemd/system/vps-inspector.service
```

它会指向：

```text
/vps-control-panel/systemd/vps-inspector.service
```

## 一键卸载

删除服务、二进制文件和配置：

```bash
curl -fsSL https://raw.githubusercontent.com/huanghao256/vps/main/scripts/uninstall.sh | sudo sh
```

如果想保留 `/vps-control-panel/config` 配置目录：

```bash
KEEP_CONFIG=1 curl -fsSL https://raw.githubusercontent.com/huanghao256/vps/main/scripts/uninstall.sh | sudo -E sh
```

## 本地开发

启动 Go 后端：

```bash
go run ./cmd/vps-inspector
```

打开：

```text
http://127.0.0.1:8719
```

构建前端：

```bash
cd web
npm install
npm run build
```

构建后端二进制：

```bash
go build -trimpath -ldflags="-s -w" -o bin/vps-inspector ./cmd/vps-inspector
```

## 远程访问

如果需要从公网访问面板，请显式设置监听地址和强随机 Token：

```bash
VPS_INSPECTOR_ADDR=0.0.0.0:8719 VPS_INSPECTOR_AUTH_TOKEN=your-long-random-token go run ./cmd/vps-inspector
```

API 请求需要携带：

```text
Authorization: Bearer your-long-random-token
```

## 项目结构

```text
.codex/skills/          面向 AI 协作开发的 Codex skill
cmd/vps-inspector/       应用入口
internal/agent/          检测任务编排和运行生命周期
internal/checks/         VPS 质量检测项
internal/config/         环境变量配置
internal/firewall/       防火墙后端识别和端口规则操作
internal/httpapi/        HTTP 路由、中间件和处理器
internal/runner/         安全命令执行边界
internal/status/         Linux 实时系统状态采集
web/                     React + TypeScript 前端
docs/                    架构和安全文档
scripts/                 安装和卸载脚本
```

## AI 开发 Skill

仓库内置了一个 Codex skill，方便后续使用 AI 继续迭代时保持代码风格和架构边界：

```text
.codex/skills/vps-inspector-dev
```

让 AI 修改项目时可以这样说：

```text
Use $vps-inspector-dev to implement this VPS Inspector change while preserving architecture, security, and validation standards.
```

这个 skill 约束了后端包边界、前端模块拆分、Linux-only 规则、防火墙安全规则、一键安装/卸载行为和验证命令，可以减少后续迭代变成难维护代码的风险。

## 安全默认值

- 默认监听 `127.0.0.1:8719`
- 绑定公网地址时必须设置强 Token
- 检测逻辑不会执行用户传入的任意 shell 字符串
- 防火墙操作只接受端口号和 `tcp` / `udp` 协议
- 防火墙启停和端口规则修改通常需要 root 权限
- 长时间检测任务使用 context 超时控制

更多安全设计见 [docs/security.md](docs/security.md)。

## 常用命令

查看服务状态：

```bash
systemctl status vps-inspector
```

查看日志：

```bash
journalctl -u vps-inspector -f
```

重启服务：

```bash
systemctl restart vps-inspector
```

## License

MIT
