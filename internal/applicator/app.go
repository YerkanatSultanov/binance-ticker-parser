package app

import (
	"binance-ticker-parser/internal/config"
	"binance-ticker-parser/internal/logging"
	"binance-ticker-parser/internal/workers"
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

func Run() error {
	logger := logging.GetLogger()

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	maxWorkers := cfg.MaxWorkers
	numCPU := runtime.NumCPU()
	if maxWorkers > numCPU {
		maxWorkers = numCPU
	}

	var wg sync.WaitGroup
	workerManager := workers.NewWorkerManager(maxWorkers, cfg.Symbols)

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(worker *workers.Worker) {
			defer wg.Done()
			worker.Run()
		}(workerManager.Workers[i])
	}

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			totalRequests := workerManager.GetTotalRequests()
			logger.Printf("workers requests total: %d\n", totalRequests)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			text, _ := reader.ReadString('\n')
			if strings.TrimSpace(text) == "STOP" {
				quit <- syscall.SIGTERM
				return
			}
		}
	}()

	<-quit
	logger.Println("Shutting down gracefully...")
	ticker.Stop()
	workerManager.StopAll()
	wg.Wait()
	logger.Println("All workers stopped. Exiting.")
	return nil
}
