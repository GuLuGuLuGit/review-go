## review-go

A Git staged area code review tool based on LLM, supporting OpenAI, DeepSeek, Qwen (Tongyi Qianwen), and any OpenAI-compatible interfaces.

> **Language / 语言**: English | [中文](README.md)

### Features

- **Staged Area Review**: Only reviews changes after `git add`, avoiding noise.
- **Terminal TUI Interface**: Built with Bubble Tea, user-friendly.
- **Multi-Provider Support**: Switch between OpenAI / DeepSeek / Qwen or custom compatible services via configuration.
- **Security Focus**: Highlights potential security issues, error handling, and performance concerns during review.

## Installation & Build

### Prerequisites

- **Go Version**: Go 1.23+ (check `go.mod` for exact version)
- **Git**: Required to use in a Git repository

### Get Source Code

```bash
git clone https://github.com/GuLuGuLuGit/review-go.git
cd review-go
```

> **Note**: Replace `yourname` with your actual GitHub username or organization name, and update the `module` declaration in `go.mod` and import paths in code.

### Build Executable

```bash
go build -o review-go .
```

After building, `review-go` (or `review-go.exe` on Windows) will be generated in the current directory.

## Configuration

review-go reads configuration from `~/.review-go.yaml` in your home directory:

- Linux / macOS: `~/.review-go.yaml`
- Windows: `%USERPROFILE%\.review-go.yaml`

A template file `config.example.yaml` is provided in the repository. You can create your own configuration based on it.

### Using Configuration Template

There are two ways to configure the project:

#### Method A: Using Command Line Tool (Recommended)

Use the `review-go config` command for quick configuration:

```bash
# Set DeepSeek API Key
review-go config set-key sk-your-api-key-here --provider deepseek

# Set OpenAI API Key
review-go config set-key sk-your-api-key-here --provider openai

# Set Qwen API Key
review-go config set-key sk-your-api-key-here --provider qwen

# Switch default provider
review-go config set-provider deepseek
```

#### Method B: Manual Configuration File Editing

1. **Copy Template File**

   ```bash
   # Linux / macOS
   cp config.example.yaml ~/.review-go.yaml

   # Windows (PowerShell)
   Copy-Item config.example.yaml "$env:USERPROFILE\.review-go.yaml"
   ```

2. **Edit Configuration File**: Open `~/.review-go.yaml`, fill in your real `api_key` based on the LLM service you use, and optionally modify `base_url` and `model`.

3. **Select Default Provider**: Set the `provider` field to your preferred provider, for example:

   ```yaml
   provider: "deepseek"  # openai / deepseek / qwen / custom name
   ```

### Multi-Provider Configuration (Recommended)

`config.example.yaml` provides a multi-provider configuration example. The core structure is as follows:

```yaml
provider: "deepseek"  # Current default provider

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

> **Important**: The repository does not contain any real keys. All examples are placeholders. Do not commit `~/.review-go.yaml` containing real `api_key`.

### Simple Configuration (Backward Compatible)

If you only need one provider, you can use the simple configuration (no `providers` field required):

```yaml
api_key: "sk-your-api-key-here"
# base_url: ""      # Optional
# model: "gpt-4o-mini"  # Optional
```

In this mode:

- `provider` can be empty
- `api_key` is required
- `base_url` and `model` will be auto-filled based on known vendors or default models

## Usage

### Configuration Commands

review-go provides convenient configuration commands:

```bash
# View all configuration commands
review-go config --help

# Set API Key for a specific provider
review-go config set-key <api-key> --provider <provider-name>

# Set default provider
review-go config set-provider <provider-name>
```

**Examples**:

```bash
# Set DeepSeek API Key
review-go config set-key sk-xxxxxxxx --provider deepseek

# Set OpenAI API Key
review-go config set-key sk-xxxxxxxx --provider openai

# Switch default provider to deepseek
review-go config set-provider deepseek
```

### Basic Workflow

1. Make code changes in your project
2. Stage the changes you want to review using Git:

   ```bash
   git add path/to/your_file.go
   ```

3. Run review-go in the repository root:

   ```bash
   ./review-go       # Linux / macOS
   .\review-go.exe  # Windows
   ```

4. The terminal will launch a TUI interface, read the Go code diff from the current staged area, send it to the configured LLM for review, and display the review results in Markdown format.

### Review Focus Areas

The review prompt focuses on:

- **Security**: Input validation, sensitive information handling, concurrency safety, etc.
- **Error Handling**: Whether errors are ignored, whether error messages are clear, whether proper error wrapping is used
- **Performance & Resource Usage**: Algorithm complexity, memory allocation, I/O patterns, potential bottlenecks, etc.

## Security & Privacy

- **Key Storage**: All API Keys are only stored locally in `~/.review-go.yaml` and will not be written to the repository.
- **Repository Security**:
  - `.gitignore` ignores all `*.yaml` configuration files by default, but keeps `config.example.yaml` as a template.
  - Do not commit configuration files containing real keys to the Git repository.
- **Code Content**: The tool sends staged diff content to your configured LLM service. Please decide whether to use it based on your team/company security policies.

## Development & Contribution

- **Module Path**:
  - In your own repository, replace `module github.com/GuLuGuLuGit/review-go` in `go.mod` with your actual repository address.
  - Also update import paths in code (`github.com/GuLuGuLuGit/review-go/...`).
- **PRs Welcome**:
  - Add more LLM provider adapters
  - Optimize TUI experience
  - Improve prompts and review report format

## License

Fill in according to your chosen open source license (e.g., MIT / Apache-2.0, etc.).
