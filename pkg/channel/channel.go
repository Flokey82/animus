package channel

import (
	"context"
	"sync"
	"time"
)

// Message encapsulates a communication exchange with an agent.
type Message struct {
	Source    string    `json:"source"`    // "user", "agent", "system"
	Channel   string    `json:"channel"`   // "cli", "discord", "telegram", "web"
	Sender    string    `json:"sender"`    // user name / handle
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// InputChannel provides incoming messages from users or systems.
type InputChannel interface {
	Name() string
	Read(ctx context.Context) (Message, error)
}

// OutputChannel transmits agent responses to an external destination.
type OutputChannel interface {
	Name() string
	Send(ctx context.Context, msg Message) error
}

// Hub manages multi-channel broadcasts and listeners.
type Hub struct {
	mu        sync.RWMutex
	listeners []chan Message
}

// NewHub creates an event broadcast hub.
func NewHub() *Hub {
	return &Hub{
		listeners: make([]chan Message, 0),
	}
}

// Subscribe returns a channel that receives all broadcast messages.
func (h *Hub) Subscribe() <-chan Message {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan Message, 20)
	h.listeners = append(h.listeners, ch)
	return ch
}

// Broadcast sends a message to all active subscribers.
func (h *Hub) Broadcast(msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, ch := range h.listeners {
		select {
		case ch <- msg:
		default:
			// avoid blocking if listener is slow
		}
	}
}
