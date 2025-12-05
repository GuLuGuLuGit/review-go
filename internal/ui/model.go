package ui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/GuLuGuLuGit/review-go/internal/ai"
	"github.com/GuLuGuLuGit/review-go/internal/gitops"
)

// reviewLoadedMsg 是后台审核任务完成后发送给 UI 的消息。
type reviewLoadedMsg struct {
	files   []string
	reviews map[string]string
	err     error
}

// Model 是 Bubble Tea 的主状态机。
//
// - files: 暂存区中有变更的文件列表
// - reviews: 每个文件对应的 LLM 审查结果（Markdown）
// - loading: 是否处于加载状态（调用 Git + AI 中）
// - selected: 当前选中的文件索引
// - err: 加载过程中的错误（如果有）
// - provider: 用于实际调用 LLM 的接口实现
type Model struct {
	files    []string
	reviews  map[string]string
	loading  bool
	selected int

	spinner  spinner.Model
	width    int
	height   int
	provider ai.LLMProvider

	err error
}

// 一些简单的样式定义，使用 lipgloss。
var (
	spinnerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Padding(1, 2)

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Padding(0, 2)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		Padding(1, 2)

	fileListStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	selectedFileStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)

	normalFileStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	reviewStyle = lipgloss.NewStyle().
		Padding(0, 1)
)

// NewModel 创建一个带有初始 loading 状态和 Spinner 的 Model。
// 通过依赖注入的方式传入一个实现了 LLMProvider 接口的实例，
// 方便后续在不同 AI 提供商之间切换。
func NewModel(provider ai.LLMProvider) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return Model{
		files:    nil,
		reviews:  make(map[string]string),
		loading:  true,
		selected: 0,
		spinner:  s,
		provider: provider,
	}
}

// Init 在程序启动时被调用，这里启动：
// 1. spinner 的 Tick
// 2. 后台 Git + AI 审核任务
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadReviewsCmd(m.provider),
	)
}

// loadReviewsCmd 在后台执行 Git + AI 审核逻辑，完成后发送 reviewLoadedMsg。
func loadReviewsCmd(provider ai.LLMProvider) tea.Cmd {
	return func() tea.Msg {
		if provider == nil {
			return reviewLoadedMsg{err: fmt.Errorf("LLM Provider 未初始化")}
		}

		files, err := gitops.GetChangedFiles()
		if err != nil {
			return reviewLoadedMsg{err: fmt.Errorf("获取暂存区文件失败：%w", err)}
		}

		if len(files) == 0 {
			return reviewLoadedMsg{
				files:   []string{},
				reviews: map[string]string{},
				err:     nil,
			}
		}

		reviews := make(map[string]string, len(files))
		for _, f := range files {
			diff, err := getFileStagedDiff(f)
			if err != nil {
				return reviewLoadedMsg{
					err: fmt.Errorf("获取文件 %s 的 diff 失败：%w", f, err),
				}
			}

			// 组合审查提示词，将原先 Reviewer 中的系统提示融合到单条 prompt 中，
			// 通过 LLMProvider 的 Chat 方法调用。
			reviewPrompt := buildReviewPrompt(diff)

			reply, err := provider.Chat(reviewPrompt)
			if err != nil {
				return reviewLoadedMsg{
					err: fmt.Errorf("审查文件 %s 失败：%w", f, err),
				}
			}

			reviews[f] = reply
		}

		return reviewLoadedMsg{
			files:   files,
			reviews: reviews,
			err:     nil,
		}
	}
}

// getFileStagedDiff 获取单个文件在暂存区中的 diff（仅该文件），等价于：
// git diff --cached --unified=0 -- <file>
func getFileStagedDiff(file string) (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--unified=0", "--", file)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		if strings.Contains(output, "not a git repository") {
			return "", fmt.Errorf("当前目录不是 git 仓库：%s", output)
		}
		if output != "" {
			return "", fmt.Errorf("执行 git diff 失败：%s", output)
		}
		return "", fmt.Errorf("执行 git diff 失败：%w", err)
	}

	if output == "" {
		return "", fmt.Errorf("文件 %s 在暂存区没有 diff 输出", file)
	}

	return output, nil
}

// Update 处理所有消息（键盘事件、窗口大小变化、后台任务结果等）。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case reviewLoadedMsg:
		m.loading = false
		m.err = msg.err

		if msg.err == nil {
			m.files = msg.files
			m.reviews = msg.reviews
			if len(m.files) > 0 && m.selected >= len(m.files) {
				m.selected = 0
			}
		}
		return m, nil

	case tea.KeyMsg:
		// 全局退出快捷键
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

		if m.loading || m.err != nil {
			// 加载中或出错时，不处理上下选择
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.files)-1 {
				m.selected++
			}
		}
	}

	return m, nil
}

// View 根据当前状态渲染 TUI。
func (m Model) View() string {
	if m.loading {
		return m.viewLoading()
	}

	if m.err != nil {
		return m.viewError()
	}

	return m.viewContent()
}

func (m Model) viewLoading() string {
	sp := m.spinner.View()
	text := "正在分析暂存区中的 Go 代码并调用 AI 进行审查，请稍候...\n(按 q 退出)"

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		spinnerStyle.Render(sp),
		infoStyle.Render(text),
	)

	return centerInTerminal(content, m.width, m.height)
}

func (m Model) viewError() string {
	msg := fmt.Sprintf("发生错误：\n\n%s\n\n按 q 退出。", m.err.Error())
	content := errorStyle.Render(msg)
	return centerInTerminal(content, m.width, m.height)
}

func (m Model) viewContent() string {
	if len(m.files) == 0 {
		msg := "暂存区中没有 .go 文件的变更。\n\n请在 Git 暂存区中添加一些 Go 代码的修改后重新运行。\n\n按 q 退出。"
		return centerInTerminal(infoStyle.Render(msg), m.width, m.height)
	}

	// 简单的左右布局：左侧文件列表，右侧审查内容
	totalWidth := m.width
	if totalWidth <= 0 {
		totalWidth = 100
	}

	leftWidth := totalWidth / 4
	if leftWidth < 20 {
		leftWidth = 20
	}

	rightWidth := totalWidth - leftWidth - 4
	if rightWidth < 20 {
		rightWidth = 20
	}

	// 构造文件列表
	var fileLines []string
	for i, f := range m.files {
		line := f
		if i == m.selected {
			line = selectedFileStyle.Render("> " + f)
		} else {
			line = normalFileStyle.Render("  " + f)
		}
		fileLines = append(fileLines, line)
	}

	fileList := strings.Join(fileLines, "\n")
	fileListBox := fileListStyle.
		Width(leftWidth).
		Render(fileList)

	// 当前选中文件对应的审查结果
	var reviewMD string
	if m.selected >= 0 && m.selected < len(m.files) {
		file := m.files[m.selected]
		reviewMD = m.reviews[file]
		if strings.TrimSpace(reviewMD) == "" {
			reviewMD = "_该文件暂无审查结果。_"
		}
	} else {
		reviewMD = "_未选中文件。_"
	}

	renderedReview := renderMarkdown(reviewMD, rightWidth)
	reviewBox := reviewStyle.
		Width(rightWidth).
		Render(renderedReview)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		fileListBox,
		reviewBox,
	)
}

// renderMarkdown 使用 glamour 渲染 Markdown 内容，如果失败则退回原始文本。
func renderMarkdown(md string, width int) string {
	if width <= 0 {
		width = 80
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return md
	}

	out, err := r.Render(md)
	if err != nil {
		return md
	}

	return out
}

// centerInTerminal 尝试把内容在终端中居中显示（如果知道终端宽高）。
func centerInTerminal(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return content
	}

	box := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)

	return box
}

// buildReviewPrompt 根据 Git diff 构造发送给 LLM 的审查提示词。
// 这里复用原先 Reviewer 中的系统说明，只是将其合并为一条用户消息，
// 以便通过通用的 Chat 接口发送。
func buildReviewPrompt(diff string) string {
	diff = strings.TrimSpace(diff)
	if diff == "" {
		return "暂存区 diff 为空，无需审查。"
	}

	systemPrompt := `你是一名资深 Golang 专家，擅长设计高可读性、可维护且鲁棒的 Go 代码。
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
		"请审查以下 Git diff（只读即可，不需要给出可直接应用的 patch），并按照上述要求返回 Markdown 格式的审查报告：\n\n```diff\n%s\n```",
		diff,
	)

	return systemPrompt + "\n\n" + userPrompt
}
