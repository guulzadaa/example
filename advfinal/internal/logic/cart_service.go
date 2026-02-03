package logic

import (
	"bookstore/internal/models"
	"fmt"
	"time"
)

type CartTask struct {
	Type string
	Item models.CartItem
}

var CartJobQueue = make(chan CartTask, 100)

type CartService struct{}

func (s *CartService) AddItemToCart(item models.CartItem) {
	fmt.Printf("[CART] Item %d added to Cart %d\n", item.BookID, item.CartID)

	CartJobQueue <- CartTask{
		Type: "RESERVE_STOCK",
		Item: item,
	}
}

func StartCartWorkerPool(workerCount int) {
	for i := 1; i <= workerCount; i++ {
		go func(workerID int) {
			fmt.Printf("Worker %d ready to process cart tasks\n", workerID)
			for job := range CartJobQueue {
				processCartJob(workerID, job)
			}
		}(i)
	}
}

func processCartJob(workerID int, job CartTask) {
	fmt.Printf("[WORKER %d] Checking stock for Book ID %d...\n", workerID, job.Item.BookID)
	time.Sleep(2 * time.Second)
	fmt.Printf("[WORKER %d] Stock reserved for Book ID %d\n", workerID, job.Item.BookID)
}
