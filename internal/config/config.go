package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ProviderConfig 描述单个 LLM 提供商的配置。
//
// YAML 结构示例：
//
//	providers:
//	  openai:
//	    api_key: "sk-..."
//	    model: "gpt-4o"
//	  deepseek:
//	    api_key: "sk-..."
//	    base_url: "https://api.deepseek.com"
//	    model: "deepseek-coder"
//	  qwen:
//	    api_key: "sk-..."
//	    base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
//	    model: "qwen-turbo"
type ProviderConfig struct {
	APIKey  string `mapstructure:"api_key" yaml:"api_key"`
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`
	Model   string `mapstructure:"model" yaml:"model"`
}

// Config 保存从配置文件加载的全局配置。
//
// 期望的配置结构示例（~/.review-go.yaml）：
//
//	provider: "deepseek" # 当前默认使用的提供商: openai, deepseek, qwen (通义千问)
//
//	providers:
//	  openai:
//	    api_key: "sk-..."
//	    model: "gpt-4o"
//	  deepseek:
//	    api_key: "sk-..."
//	    base_url: "https://api.deepseek.com"
//	    model: "deepseek-coder"
//	  qwen:
//	    api_key: "sk-..."
//	    base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
//	    model: "qwen-turbo"
//
// 同时，为了兼容之前只有 api_key 的简单配置：
//
//	api_key: "sk-xxxxx"
//
// 会被视为单一默认提供商。
type Config struct {
	// Provider 是当前默认使用的提供商名称（如："openai"、"deepseek"、"qwen"）。
	Provider string `mapstructure:"provider" yaml:"provider"`

	// Providers 是一个以提供商名称为 key 的配置映射。
	Providers map[string]ProviderConfig `mapstructure:"providers" yaml:"providers"`

	// 以下字段为“当前激活提供商”的扁平配置，方便其他模块直接使用。
	// 如果配置文件使用多提供商结构，则这些字段会在 Load() 中
	// 根据 provider 自动从 providers[provider] 中填充。
	//
	// 如果配置文件仍然是旧版结构，仅包含 api_key，则：
	// - Provider 为空
	// - Providers 为空
	// - 仅 APIKey 有值，保持向后兼容。
	APIKey  string `mapstructure:"api_key" yaml:"api_key"`
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`
	Model   string `mapstructure:"model" yaml:"model"`
}

// Load 从 ~/.review-go.yaml 读取配置。
//
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get user home dir: %w", err)
	}

	configPath := filepath.Join(home, ".review-go.yaml")

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 提前设置默认值，避免 key 不存在时返回空字符串。
	// 对于多提供商结构，我们只给最顶层 provider 设置一个空默认值，
	// 实际必填校验在后面进行。
	v.SetDefault("provider", "")
	v.SetDefault("api_key", "")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file %s: %w", configPath, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// 如果配置中定义了 providers，则走多提供商逻辑。
	if len(cfg.Providers) > 0 {
		if cfg.Provider == "" {
			return nil, fmt.Errorf("provider is empty in %s", configPath)
		}

		providerCfg, ok := cfg.Providers[cfg.Provider]
		if !ok {
			return nil, fmt.Errorf("provider %q not found under providers in %s", cfg.Provider, configPath)
		}

		if providerCfg.APIKey == "" {
			return nil, fmt.Errorf("api_key for provider %q is empty in %s", cfg.Provider, configPath)
		}

		// 将当前 provider 的配置“扁平化”到顶层字段，方便其他模块直接使用。
		cfg.APIKey = providerCfg.APIKey
		cfg.BaseURL = providerCfg.BaseURL
		cfg.Model = providerCfg.Model

		return &cfg, nil
	}

	// 兼容旧版：没有 providers 字段时，仍然要求存在顶层 api_key。
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is empty in %s", configPath)
	}

	return &cfg, nil
}


