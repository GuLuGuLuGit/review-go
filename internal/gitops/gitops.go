package gitops

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// GetStagedDiff 返回当前 Git 仓库中暂存区（index）里所有 .go 文件的 diff。
//
// 实现等价于在命令行执行：
//
//	git diff --cached --unified=0 -- *.go
//
// 仅返回标准输出内容，如果 git 未安装、当前目录不是 git 仓库、或命令执行失败，
// 会返回带有清晰信息的错误。
func GetStagedDiff() (string, error) {
	args := []string{"diff", "--cached", "--unified=0", "--", "*.go"}
	cmd := exec.Command("git", args...)

	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("git 未安装或不在 PATH 中: %w", err)
		}

		if strings.Contains(output, "not a git repository") {
			return "", fmt.Errorf("当前目录不是 git 仓库: %s", output)
		}

		if output != "" {
			return "", fmt.Errorf("执行 git diff 失败: %s", output)
		}

		return "", fmt.Errorf("执行 git diff 失败: %w", err)
	}

	return output, nil
}

// GetChangedFiles 返回暂存区中有变更的 .go 文件列表。
//
// 实现等价于在命令行执行：
//
//	git diff --cached --name-only -- *.go
//
// 返回去重且非空的文件路径切片（相对于仓库根目录）。
func GetChangedFiles() ([]string, error) {
	args := []string{"diff", "--cached", "--name-only", "--", "*.go"}
	cmd := exec.Command("git", args...)

	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("git 未安装或不在 PATH 中: %w", err)
		}

		if strings.Contains(output, "not a git repository") {
			return nil, fmt.Errorf("当前目录不是 git 仓库: %s", output)
		}

		if output != "" {
			return nil, fmt.Errorf("执行 git diff --name-only 失败: %s", output)
		}

		return nil, fmt.Errorf("执行 git diff --name-only 失败: %w", err)
	}

	if output == "" {
		// 暂存区没有 .go 文件的变更，返回空切片而不是 nil，方便调用方直接 range
		return []string{}, nil
	}

	lines := strings.Split(output, "\n")
	files := make([]string, 0, len(lines))
	seen := make(map[string]struct{}, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		files = append(files, line)
	}

	return files, nil
}


