package main

import (
	"log"

	"github.com/yourname/review-go/cmd"
	"github.com/yourname/review-go/internal/config"
)

func main() {
	// 预先尝试加载配置，主要是验证 ~/.review-go.yaml 是否可用
	_, err := config.Load()
	if err != nil {
		// 配置读取失败并不阻止 CLI 运行，只做提示
		log.Printf("warning: failed to load config (~/.review-go.yaml): %v", err)
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}


