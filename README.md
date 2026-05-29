
# wg-publickey-prefix-gen
一个用于本地生成符合特定 Base64 前缀的 X25519（WireGuard）公钥的工具集合，包含 Go 和 Python 两种实现。

**主要特性**
- 支持多线程并行搜索（默认使用 CPU 核心数）。
- 输出 Base64 编码的私钥与公钥，实时显示尝试次数与速度监控。
- 可配置线程数与需要找到的结果数量（count）。

**仓库结构**
- `go/` — Go 实现（高性能，使用非加密伪随机为每个 worker 加速）。
- `py/` — Python 实现（基于 `cryptography`，跨平台）。
- `Makefile`, `README.md`, `LICENSE` — 项目根级常用文件。
- `.github/workflows/go-cross-compile.yml` — 手动触发的 Go 交叉编译 workflow。

**最低要求**
- Go: `go` 1.18+（构建/运行 Go 实现时需要）。
- Python: `python3` 和 `cryptography` 包（可用 `pip install cryptography` 安装）。

使用说明
---------

**Go 实现**  
- 使用 Makefile 构建当前架构：

```bash
make build
```

- 单平台构建目标：

```bash
make go-build-linux-amd64
make go-build-linux-arm64
make go-build-windows-amd64
make go-build-darwin-amd64
make go-build-darwin-arm64
```

- 完整交叉编译：

```bash
make go-full-build
```

- 产物会输出到 `dist/`

- 运行示例：

```bash
./wg_vanity <prefix> [threads] [count]
# 例: ./wg_vanity abc 8 2   -> 搜索以 "abc" 开头的公钥，用 8 线程，找到 2 个结果后退出
```

- 参数说明：
	- `<prefix>`: 要匹配的 Base64 公钥前缀（必填）。
	- `[threads]`: 可选，工作线程数，默认为 CPU 核心数。
	- `[count]` : 可选，需要找到的结果数量，默认 1。

**Python 实现**
- 安装依赖：

```bash
pip install cryptography
```

- 运行示例：

```bash
./wg-vanity.py <prefix> [threads] [count]
# Linux/macOS 可直接执行；Windows 请使用：python3 py/wg-vanity.py <prefix> [threads] [count]
```

- 如果你希望显式调用解释器，也可以直接运行：

```bash
python3 py/wg-vanity.py <prefix> [threads] [count]
```

**输出与行为**
- 两个实现都会在找到结果时打印 `FOUND #N!`、线程号、私钥与公钥（Base64），并打印尝试总次数与速度指标。
- 会在退出时显示已尝试的总次数与已找到的数量。

**性能提示**
- 更长或更具体的前缀显著增加搜索时间；可通过增加线程数或在更强的硬件上运行来提升速度。
- Go 实现使用每-worker 的伪随机以换取更高吞吐量；Python 实现更易读、便于修改。

许可
- [CC0-1.0](LICENSE)