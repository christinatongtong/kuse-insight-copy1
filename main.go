package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kuse-ai/kuse-insight-go/inputs"
	"github.com/kuse-ai/kuse-insight-go/insights"
	"github.com/kuse-ai/kuse-insight-go/llm"
	"github.com/kuse-ai/kuse-insight-go/logger"
	"github.com/kuse-ai/kuse-insight-go/outputs"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load() // 默认加载 .env
	if err != nil {
		log.Fatal()
	}
}

func main() {
	insight := insights.NewUserInsight(
		inputs.NewInputs(),
		outputs.NewOutputs(),
		insights.WithModel(llm.NewGPT4Dot1Model()),
		insights.WithMaxConcurrency(100),
	)
	go StartSignalHandler(insight)
	// insight.Run(context.TODO(), "742")
	// insight.RunAll() // ./sources/big_query/users.csv数据跑全量
	// insight.Cluster() // 对results/results.csv进行聚类
	insight.UploadMixpanel() // 把results/results.csv上传到mixpanel
	logger.Info("All Finished")
}

func StartSignalHandler(insight *insights.UserInsights) {
	_, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGSEGV)
	sig := <-sigChan

	switch sig {
	case syscall.SIGQUIT,
		syscall.SIGABRT,
		syscall.SIGSEGV,
		syscall.SIGINT:
		insight.Save()
	}

	cancel()

	go func() {
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
}
