package workerpool

import (
	"log"

	"github.com/joshbarros/golang-chat-api/internal/domain"
	"github.com/joshbarros/golang-chat-api/internal/repository"
)

type WorkerPool struct {
	jobQueue    chan domain.Message
	messageRepo *repository.MessageRepository
}

func NewWorkerPool(numWorkers int, messageRepo *repository.MessageRepository) *WorkerPool {
	wp := &WorkerPool{
		jobQueue:    make(chan domain.Message, 100), // Queue size of 100
		messageRepo: messageRepo,
	}

	// Launch workers
	for i := 0; i < numWorkers; i++ {
		go wp.worker(i)
	}
	return wp
}

func (wp *WorkerPool) worker(id int) {
	// Workers listen on the global jobQueue
	for msg := range wp.jobQueue {
		// Log the message before processing
		log.Printf("Worker %d processing message from user %d in room %s: %s", id, msg.UserID, msg.RoomID, msg.Message)

		// Save the message to the database
		if err := wp.messageRepo.SaveMessage(msg); err != nil {
			log.Printf("Worker %d failed to save message from user %d in room %s: %v", id, msg.UserID, msg.RoomID, err)
		} else {
			log.Printf("Worker %d successfully saved message from user %d in room %s", id, msg.UserID, msg.RoomID)
		}
	}
}

func (wp *WorkerPool) AddJob(msg domain.Message) {
	wp.jobQueue <- msg
	log.Printf("Job added to worker pool for user %d in room %s", msg.UserID, msg.RoomID)
}
