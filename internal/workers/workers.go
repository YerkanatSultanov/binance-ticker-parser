package workers

import (
	"binance-ticker-parser/pkg/binance"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Worker struct {
	Symbols        []string
	RequestCount   int
	previousPrices map[string]string
	mu             sync.Mutex
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewWorker(symbols []string) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		Symbols:        symbols,
		previousPrices: make(map[string]string),
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (w *Worker) Run() {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			for _, symbol := range w.Symbols {
				price, err := binance.GetPrice(symbol)
				if err != nil {
					log.Printf("Error fetching price for %s: %v", symbol, err)
					continue
				}
				w.mu.Lock()
				w.RequestCount++
				changed := ""
				if prevPrice, ok := w.previousPrices[symbol]; ok && prevPrice != price {
					changed = "changed"
				}
				w.previousPrices[symbol] = price
				w.mu.Unlock()
				fmt.Printf("%s price:%s %s\n", symbol, price, changed)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (w *Worker) GetRequestsCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.RequestCount
}

func (w *Worker) Stop() {
	w.cancel()
}

type WorkerManager struct {
	Workers []*Worker
}

func NewWorkerManager(maxWorkers int, symbols []string) *WorkerManager {
	workers := make([]*Worker, maxWorkers)
	symbolsPerWorker := len(symbols) / maxWorkers

	for i := 0; i < maxWorkers; i++ {
		start := i * symbolsPerWorker
		end := start + symbolsPerWorker
		if i == maxWorkers-1 {
			end = len(symbols)
		}
		workers[i] = NewWorker(symbols[start:end])
	}

	return &WorkerManager{Workers: workers}
}

func (wm *WorkerManager) GetTotalRequests() int {
	totalRequests := 0
	for _, w := range wm.Workers {
		totalRequests += w.GetRequestsCount()
	}
	return totalRequests
}

func (wm *WorkerManager) StopAll() {
	for _, w := range wm.Workers {
		w.Stop()
	}
}
