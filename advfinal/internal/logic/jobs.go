package logic

import (
	"log"

	"bookstore/internal/repository"
)

type OrderJobType string

const (
	JobAuditOrderCreated OrderJobType = "AUDIT_ORDER_CREATED"
	JobClearCart         OrderJobType = "CLEAR_CART"
	JobClearWishlist     OrderJobType = "CLEAR_WISHLIST"
)

type OrderJob struct {
	Type       OrderJobType
	OrderID    int
	CartID     int
	WishlistID int
}

var OrderJobQueue = make(chan OrderJob, 100)

func StartOrderWorkerPool(workerCount int, cartRepo repository.CartRepository, wishlistRepo repository.WishlistRepository) {
	log.Printf("[ORDER WORKERS] starting %d workers...\n", workerCount)

	for i := 1; i <= workerCount; i++ {
		go func(workerID int) {
			log.Printf("[ORDER WORKER %d] started\n", workerID)

			for job := range OrderJobQueue {
				log.Printf("[ORDER WORKER %d] got job: type=%s orderId=%d cartId=%d wishlistId=%d\n",
					workerID, job.Type, job.OrderID, job.CartID, job.WishlistID)

				switch job.Type {
				case JobClearCart:
					if err := cartRepo.ClearCart(job.CartID); err != nil {
						log.Printf("[ORDER WORKER %d] ClearCart failed: %v\n", workerID, err)
					} else {
						log.Printf("[ORDER WORKER %d] cart cleared: cartId=%d\n", workerID, job.CartID)
					}

				case JobClearWishlist:
					if err := wishlistRepo.Delete(job.WishlistID); err != nil {
						log.Printf("[ORDER WORKER %d] Delete wishlist failed: %v\n", workerID, err)
					} else {
						log.Printf("[ORDER WORKER %d] wishlist cleared: wishlistId=%d\n", workerID, job.WishlistID)
					}
				case JobAuditOrderCreated:
					log.Printf("[ORDER WORKER %d] audit: order created orderId=%d\n", workerID, job.OrderID)

				default:
					log.Printf("[ORDER WORKER %d] unknown job type: %s\n", workerID, job.Type)
				}
			}
		}(i)
	}
}
