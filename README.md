## review-go

一个基于 LLM 的 Git 暂存区代码审查工具，支持 OpenAI、DeepSeek、通义千问（Qwen）以及任意 OpenAI 兼容接口。

> **Language / 语言**: [English](README_EN.md) | 中文

### 功能特性

- **基于暂存区审查**：只审查 `git add` 后的变更，避免噪音。
- **终端 TUI 界面**：基于 Bubble Tea，交互友好。
- **多提供商支持**：通过配置切换 OpenAI / DeepSeek / Qwen 或自定义兼容服务。
- **安全关注点**：审查中重点提示潜在安全问题、错误处理与性能隐患。

## 安装与构建

### 前置条件

- **Go 版本**: Go \(推荐 1.23+，以 `go.mod` 为准\)
- **Git**: 需要在一个 Git 仓库中使用

### 获取源码

```bash
git clone https://github.com/yourname/review-go.git
cd review-go
```

> **提示**：开源时请将 `yourname` 替换为你的实际 GitHub 用户名或组织名，并同步更新 `go.mod` 中的 `module` 声明和代码中的导入路径。

### 构建可执行文件

```bash
go build -o review-go .
```

构建完成后，会在当前目录生成 `review-go`（Windows 下为 `review-go.exe`）。

## 配置说明

review-go 的配置文件默认读取自当前用户主目录下的 `~/.review-go.yaml`：

- Linux / macOS: `~/.review-go.yaml`
- Windows: `%USERPROFILE%\.review-go.yaml`

仓库中提供了一个模板文件：`config.example.yaml`，你可以基于它创建自己的配置。

### 使用配置模板

有两种方式配置项目：

#### 方式 A: 使用命令行工具（推荐）

使用 `review-go config` 命令快速配置：

```bash
# 设置 DeepSeek 的 API Key
review-go config set-key sk-your-api-key-here --provider deepseek

# 设置 OpenAI 的 API Key
review-go config set-key sk-your-api-key-here --provider openai

# 设置通义千问的 API Key
review-go config set-key sk-your-api-key-here --provider qwen

# 切换默认提供商
review-go config set-provider deepseek
```

#### 方式 B: 手动编辑配置文件

1. **复制模板文件**

   ```bash
   # Linux / macOS
   cp config.example.yaml ~/.review-go.yaml

   # Windows（PowerShell）
   Copy-Item config.example.yaml "$env:USERPROFILE\.review-go.yaml"
   ```

2. **编辑配置文件**：打开 `~/.review-go.yaml`，根据你实际使用的 LLM 服务，填入真实的 `api_key`，可选择性修改 `base_url` 和 `model`。

3. **选择默认提供商**：将 `provider` 字段设置为你想默认使用的提供商，例如：

   ```yaml
   provider: "deepseek"  # openai / deepseek / qwen / 自定义名称
   ```

### 多提供商配置（推荐）

`config.example.yaml` 中已经给出了多提供商配置示例，核心结构如下：

```yaml
provider: "deepseek"  # 当前默认使用的提供商

providers:
  openai:
    api_key: "sk-your-openai-api-key-here"
    model: "gpt-4o"

  deepseek:
    api_key: "sk-your-deepseek-api-key-here"
    base_url: "https://api.deepseek.com"
    model: "deepseek-coder"

  qwen:
    api_key: "sk-your-qwen-api-key-here"
    base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
    model: "qwen-turbo"
```

> **重要**：仓库中不会包含任何真实密钥，所有示例都是占位符。请勿提交包含真实 `api_key` 的 `~/.review-go.yaml`。

### 简单配置（向后兼容）

如果你只想使用一个提供商，也可以使用简单配置方式（无需 `providers` 字段）：

```yaml
api_key: "sk-your-api-key-here"
# base_url: ""      # 可选
# model: "gpt-4o-mini"  # 可选
```

在这种模式下：

- `provider` 可以留空
- `api_key` 必填
- `base_url` 和 `model` 会根据已知厂商或默认模型自动补全

## 使用方法

### 配置命令

review-go 提供了便捷的配置命令：

```bash
# 查看所有配置命令
review-go config --help

# 设置指定提供商的 API Key
review-go config set-key <api-key> --provider <provider-name>

# 设置默认提供商
review-go config set-provider <provider-name>
```

**示例**：

```bash
# 设置 DeepSeek API Key
review-go config set-key sk-xxxxxxxx --provider deepseek

# 设置 OpenAI API Key
review-go config set-key sk-xxxxxxxx --provider openai

# 切换默认提供商为 deepseek
review-go config set-provider deepseek
```

### 基本流程

1. 在你的项目中进行代码修改
2. 使用 Git 将需要审查的变更加入暂存区：

   ```bash
   git add path/to/your_file.go
   ```

3. 在仓库根目录执行 review-go：

   ```bash
   ./review-go       # Linux / macOS
   .\review-go.exe  # Windows
   ```

4. 终端会启动一个 TUI 界面，读取当前暂存区中的 Go 代码 diff，将其发送给配置好的 LLM 进行审查，并以 Markdown 形式展示审查结果。

### 审查内容重点

审查提示词会重点关注：

- **安全性**：输入校验、敏感信息处理、并发安全等
- **错误处理**：错误是否被忽略、错误信息是否清晰、是否有合理的 wrapping
- **性能与资源使用**：算法复杂度、内存分配、I/O 模式、可能的瓶颈等

## 安全与隐私

- **密钥存储**：所有 API Key 仅保存在本地的 `~/.review-go.yaml` 中，不会写入仓库。
- **仓库安全**：
  - `.gitignore` 已默认忽略所有 `*.yaml` 配置文件，但保留 `config.example.yaml` 作为模板。
  - 请勿将包含真实密钥的配置文件提交到 Git 仓库。
- **代码内容**：工具会将暂存区的 diff 内容发送给你配置的 LLM 服务，请根据团队/公司安全策略决定是否可用。

## 开发与贡献

- **模块路径**：
  - 请在自己的仓库中，将 `go.mod` 中的 `module github.com/yourname/review-go` 替换为你的实际仓库地址。
  - 同时更新代码中的导入路径（`github.com/yourname/review-go/...`）。
- **欢迎 PR**：
  - 新增更多 LLM 提供商适配
  - 优化 TUI 体验
  - 改进提示词与审查报告格式

## 许可协议

根据你选择的开源协议填写（例如 MIT / Apache-2.0 等）。
