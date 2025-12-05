package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/yourname/review-go/internal/ai"
	"github.com/yourname/review-go/internal/config"
	"github.com/yourname/review-go/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "review-go",
	Short: "review-go 是一个基于 LLM 的 Git 暂存区代码审查工具",
	Long: `review-go 是一个命令行工具，用于读取本地 Git 仓库暂存区的代码，
将分阶段变更发送给 LLM 进行代码审查，并在终端 TUI 中展示审查结果。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 读取配置并创建对应的 LLM Provider（支持 openai/deepseek/qwen 等）
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		provider, err := ai.NewProvider(*cfg)
		if err != nil {
			return fmt.Errorf("初始化 LLM Provider 失败: %w", err)
		}

		// 启动 Bubble Tea TUI 主界面
		m := ui.NewModel(provider)
		p := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("启动 TUI 失败: %w", err)
		}

		return nil
	},
}

// Execute 是 CLI 的入口，由 main.go 调用。
func Execute() error {
	return rootCmd.Execute()
}
