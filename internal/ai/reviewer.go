package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const (
	// 默认首选模型，可以根据实际后端替换为 "deepseek-chat" 等兼容名称。
	defaultModel = "gpt-4o-mini"

	defaultTemperature = 0.2
)

// CodeReviewer 定义了代码审查接口，方便后续在其他模块中通过接口进行依赖反转和单元测试。
type CodeReviewer interface {
	ReviewCode(diff string) (string, error)
}

// Reviewer 使用 OpenAI 兼容的 Chat 接口进行代码审查。
type Reviewer struct {
	client *openai.Client
	model  string
}

// NewReviewer 基于 API Key 创建一个 Reviewer。
//
// 如果你使用 DeepSeek 或其他 OpenAI 兼容服务，可以在外部配置 BaseURL 和自定义模型名。
func NewReviewer(apiKey string, opts ...func(*Reviewer)) *Reviewer {
	client := openai.NewClient(apiKey)

	r := &Reviewer{
		client: client,
		model:  defaultModel,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// WithModel 允许在创建 Reviewer 时覆盖默认模型，例如使用 "deepseek-chat"。
func WithModel(model string) func(*Reviewer) {
	return func(r *Reviewer) {
		if strings.TrimSpace(model) != "" {
			r.model = model
		}
	}
}

// ReviewCode 调用 LLM 对给定的 Git diff 进行代码审查，返回 Markdown 格式的审查报告。
//
// 重点关注：
//   - 安全性（输入验证、敏感信息处理、并发安全等）
//   - 错误处理（错误传播、日志、重试策略等）
//   - 性能（算法复杂度、内存分配、I/O 模式等）
func (r *Reviewer) ReviewCode(diff string) (string, error) {
	if r == nil || r.client == nil {
		return "", errors.New("Reviewer 未正确初始化：client 为空")
	}

	diff = strings.TrimSpace(diff)
	if diff == "" {
		return "", errors.New("diff 为空，无法进行代码审查")
	}

	systemPrompt := `
你是一名资深 Golang 专家，擅长设计高可读性、可维护且鲁棒的 Go 代码。
现在请你扮演“代码审查助手”，针对给定的 Git diff 进行严格的代码评审，重点关注：

1. 安全性：
   - 输入校验是否充分
   - 是否存在潜在的注入风险、越界访问、竞争条件等
   - 敏感信息（如密钥、token、密码）是否有泄露风险

2. 错误处理：
   - 错误是否被忽略或吞掉
   - 错误信息是否清晰、能帮助定位问题
   - 是否合理使用 error wrapping 以及日志

3. 性能与资源使用：
   - 算法与数据结构是否合理
   - 是否存在明显的多余分配或重复计算
   - I/O、网络、并发是否可能成为瓶颈

请以 Markdown 格式输出审查结果，建议结构示例：

## 总体评价
- 简要评价这次变更的整体质量。

## 主要风险与问题
- 按严重程度列出主要问题，并引用相关代码片段或行号（如果 diff 中有）。

## 优化建议
- 给出可以改进的地方，包括安全、错误处理和性能方面的具体建议。

## 认可的优点
- 指出本次改动中值得保留或学习的写法。

回复时只需要给出审查内容，无需重复贴出完整 diff。`

	userPrompt := fmt.Sprintf(
		"请审查以下 Git diff（只读即可，不需要给出可直接应用的 patch），并按照系统提示要求返回 Markdown 格式的审查报告：\n\n```diff\n%s\n```",
		diff,
	)

	req := openai.ChatCompletionRequest{
		Model:       r.model,
		Temperature: float32(defaultTemperature),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
	}

	ctx := context.Background()
	resp, err := r.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("调用 LLM 进行代码审查失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("LLM 返回结果为空：没有任何 choices")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	if content == "" {
		return "", errors.New("LLM 返回的审查内容为空")
	}

	return content, nil
}


