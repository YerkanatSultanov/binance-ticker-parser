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
	RequestCount   int // неплохо бы сделать это поле неимпортируемым
	previousPrices map[string]string
	mu             sync.Mutex
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewWorker(symbols []string) *Worker {
	ctx, cancel := context.WithCancel(context.Background()) // почему бы не хранить общий контекст где-то выше, чем у каждого воркера свой. Если ты потом снова запустишь Run у тебя уже контекст будет отменен, потому что ты его создал один раз в конструкторе
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
			for _, symbol := range w.Symbols { // лучше селект засунуть в это цикл - иначе будешь ждать пока по всем символам пройдешь, прежде чем выйти по контексту
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
				fmt.Printf("%s price:%s %s\n", symbol, price, changed) // тут нельзя это писать, в задаче сказано
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
	symbolsPerWorker := len(symbols) / maxWorkers // словишь панику если maxWorkers=0

	for i := 0; i < maxWorkers; i++ {
		start := i * symbolsPerWorker
		end := start + symbolsPerWorker
		if i == maxWorkers-1 {
			end = len(symbols)
		}
		workers[i] = NewWorker(symbols[start:end]) // неравномерно распределишь символы, если будет 11 символов и 7 воркеров
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
