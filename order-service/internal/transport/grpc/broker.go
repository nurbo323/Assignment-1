package grpc

import (
    "sync"
    "time"
)

type OrderStatusEvent struct {
    OrderID   string
    Status    string
    UpdatedAt time.Time
}

type Broker struct {
    mu          sync.RWMutex
    subscribers map[string]map[chan OrderStatusEvent]struct{}
}

func NewBroker() *Broker {
    return &Broker{
        subscribers: make(map[string]map[chan OrderStatusEvent]struct{}),
    }
}

func (b *Broker) Subscribe(orderID string) (chan OrderStatusEvent, func()) {
    ch := make(chan OrderStatusEvent, 1)

    b.mu.Lock()
    if _, ok := b.subscribers[orderID]; !ok {
        b.subscribers[orderID] = make(map[chan OrderStatusEvent]struct{})
    }
    b.subscribers[orderID][ch] = struct{}{}
    b.mu.Unlock()

    unsubscribe := func() {
        b.mu.Lock()
        defer b.mu.Unlock()

        if subs, ok := b.subscribers[orderID]; ok {
            delete(subs, ch)
            if len(subs) == 0 {
                delete(b.subscribers, orderID)
            }
        }
        close(ch)
    }

    return ch, unsubscribe
}

func (b *Broker) Publish(orderID string, status string) {
    event := OrderStatusEvent{
        OrderID:   orderID,
        Status:    status,
        UpdatedAt: time.Now(),
    }

    b.mu.RLock()
    defer b.mu.RUnlock()

    for ch := range b.subscribers[orderID] {
        select {
        case ch <- event:
        default:
        }
    }
}
