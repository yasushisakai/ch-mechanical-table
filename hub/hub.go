package hub

import (
	"log"
)

type BroadcastMessage struct {
	Name  string
	Value float64
}

type Hub struct {
	closeChan      chan struct{}
	DoneChan       chan struct{}
	RegisterChan   chan chan BroadcastMessage
	BroadcastChan  chan BroadcastMessage
	UnregisterChan chan chan BroadcastMessage
	clients        map[chan BroadcastMessage]struct{}
}

func New() *Hub {
	return &Hub{
		closeChan:      make(chan struct{}),
		DoneChan:       make(chan struct{}),
		RegisterChan:   make(chan chan BroadcastMessage),
		UnregisterChan: make(chan chan BroadcastMessage),
		BroadcastChan:  make(chan BroadcastMessage),
		clients:        make(map[chan BroadcastMessage]struct{}),
	}
}

func (h *Hub) Start() {
	go func() {
		defer close(h.closeChan)
		defer close(h.BroadcastChan)
		defer close(h.RegisterChan)
		defer close(h.UnregisterChan)

		for {
			select {

			case <-h.closeChan:
				h.DoneChan <- struct{}{}
				break

			case m, ok := <-h.BroadcastChan:
				if !ok {
					log.Println("Broadcast Channel closed from other end")
					return
				}
				for c := range h.clients {
					select {
					case c <- m:
					default:
					}

				}

			case cl, ok := <-h.RegisterChan:
				if !ok {
					log.Println("Register Channel closed from other end")
					return
				}
				h.clients[cl] = struct{}{}

			case cl, ok := <-h.UnregisterChan:
				if !ok {
					log.Println("Unregister Channel closed from other end")
					return
				}
				delete(h.clients, cl)
				close(cl)
			}
		}
	}()
}

func (h *Hub) Stop() {
	h.closeChan <- struct{}{}
}
