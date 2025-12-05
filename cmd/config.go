package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理配置文件",
	Long:  `管理 ~/.review-go.yaml 配置文件，包括设置 API Key、切换提供商等。`,
}

var setKeyCmd = &cobra.Command{
	Use:   "set-key",
	Short: "设置 API Key",
	Long: `设置指定提供商的 API Key。

如果使用 --provider 参数，会在多提供商配置中设置对应提供商的 API Key。
如果不使用 --provider 参数，会设置简单配置模式的 API Key。`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey := args[0]
		provider, _ := cmd.Flags().GetString("provider")

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("获取用户主目录失败: %w", err)
		}

		configPath := filepath.Join(home, ".review-go.yaml")

		// 读取现有配置
		var config map[string]interface{}
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("读取配置文件失败: %w", err)
			}

			if err := yaml.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("解析配置文件失败: %w", err)
			}
		} else {
			// 配置文件不存在，创建新的
			config = make(map[string]interface{})
		}

		// 如果指定了 provider，使用多提供商模式
		if provider != "" {
			// 确保 providers 字段存在
			if config["providers"] == nil {
				config["providers"] = make(map[string]interface{})
			}

			providers, ok := config["providers"].(map[string]interface{})
			if !ok {
				providers = make(map[string]interface{})
				config["providers"] = providers
			}

			// 获取或创建该 provider 的配置
			var providerConfig map[string]interface{}
			if existing, ok := providers[provider].(map[string]interface{}); ok {
				providerConfig = existing
			} else {
				providerConfig = make(map[string]interface{})
				providers[provider] = providerConfig
			}

			// 设置 API Key
			providerConfig["api_key"] = apiKey

			// 如果该 provider 还没有 base_url 和 model，根据已知提供商设置默认值
			if providerConfig["base_url"] == nil {
				switch provider {
				case "deepseek":
					providerConfig["base_url"] = "https://api.deepseek.com"
					if providerConfig["model"] == nil {
						providerConfig["model"] = "deepseek-coder"
					}
				case "qwen", "tongyi", "ali", "aliyun":
					providerConfig["base_url"] = "https://dashscope.aliyuncs.com/compatible-mode/v1"
					if providerConfig["model"] == nil {
						providerConfig["model"] = "qwen-turbo"
					}
				case "openai":
					// OpenAI 不需要 base_url，使用默认值
					if providerConfig["model"] == nil {
						providerConfig["model"] = "gpt-4o-mini"
					}
				}
			}

			// 如果当前没有设置默认 provider，设置为当前 provider
			if config["provider"] == nil || config["provider"] == "" {
				config["provider"] = provider
			}

			fmt.Printf("✅ 已为提供商 '%s' 设置 API Key\n", provider)
		} else {
			// 简单配置模式：直接设置顶层 api_key
			config["api_key"] = apiKey
			fmt.Println("✅ 已设置 API Key（简单配置模式）")
		}

		// 写入配置文件
		data, err := yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("序列化配置失败: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			return fmt.Errorf("写入配置文件失败: %w", err)
		}

		fmt.Printf("📝 配置文件已保存到: %s\n", configPath)
		return nil
	},
}

var setProviderCmd = &cobra.Command{
	Use:   "set-provider",
	Short: "设置默认提供商",
	Long:  `设置默认使用的 LLM 提供商。该提供商必须在 providers 配置中已存在。`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider := args[0]

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("获取用户主目录失败: %w", err)
		}

		configPath := filepath.Join(home, ".review-go.yaml")

		// 读取现有配置
		var config map[string]interface{}
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("读取配置文件失败: %w", err)
			}

			if err := yaml.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("解析配置文件失败: %w", err)
			}
		} else {
			return fmt.Errorf("配置文件不存在，请先使用 'config set-key' 设置 API Key")
		}

		// 检查 providers 是否存在
		providers, ok := config["providers"].(map[string]interface{})
		if !ok || providers == nil {
			return fmt.Errorf("未找到多提供商配置，请先使用 'config set-key --provider %s' 设置该提供商的 API Key", provider)
		}

		// 检查指定的 provider 是否存在
		if _, ok := providers[provider]; !ok {
			return fmt.Errorf("提供商 '%s' 不存在，请先使用 'config set-key --provider %s' 设置该提供商的 API Key", provider, provider)
		}

		// 设置默认 provider
		config["provider"] = provider

		// 写入配置文件
		data, err := yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("序列化配置失败: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			return fmt.Errorf("写入配置文件失败: %w", err)
		}

		fmt.Printf("✅ 已设置默认提供商为: %s\n", provider)
		fmt.Printf("📝 配置文件已保存到: %s\n", configPath)
		return nil
	},
}

func init() {
	// 添加 set-key 命令的 flag
	setKeyCmd.Flags().StringP("provider", "p", "", "提供商名称（如: openai, deepseek, qwen）")

	// 将子命令添加到 config 命令
	configCmd.AddCommand(setKeyCmd)
	configCmd.AddCommand(setProviderCmd)

	// 将 config 命令添加到根命令
	rootCmd.AddCommand(configCmd)
}
