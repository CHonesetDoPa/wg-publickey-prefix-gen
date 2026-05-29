# Python 子项目

位于此目录的 Python 实现使用系统 `wg` 命令生成 WireGuard 密钥对，是一个独立的子项目。

- 安装开发依赖并 lint：
- 运行脚本：


该脚本仅依赖 Python 标准库和系统 `wg` 命令，无需创建虚拟环境或安装额外依赖。仓库不包含任何自动 lint 或格式化步骤。

```bash
python3 wg-vanity.py <prefix> <threads>
```

```bash
python3 wg-vanity.py <prefix> <threads>
```
