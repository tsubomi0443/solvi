package sse

import (
	"encoding/json"
	"time"
)

type Event struct {
	Event string
	Data  string
}

type Client struct {
	UserID      uint
	IsSupporter bool
	Send        chan Event
}

type Hub struct {
	clients    map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = struct{}{}
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				close(c.Send)
				delete(h.clients, c)
			}
		}
	}
}

func (h *Hub) RunSSE() {
	go func() {
		tick := time.NewTicker(time.Second)
		loc, _ := time.LoadLocation("Asia/Tokyo")
		for t := range tick.C {
			h.broadcastAll(Event{Event: "time-tick", Data: t.In(loc).Format(time.RFC3339)})
		}
	}()
}

func (h *Hub) Register(c *Client) { h.register <- c }
func (h *Hub) Unregister(c *Client) { h.unregister <- c }

func (h *Hub) broadcastAll(ev Event) {
	for c := range h.clients {
		h.send(c, ev)
	}
}

func (h *Hub) send(c *Client, ev Event) {
	select {
	case c.Send <- ev:
	default:
		close(c.Send)
		delete(h.clients, c)
	}
}

func (h *Hub) SendToSupporters(event string, payload interface{}) {
	data, _ := json.Marshal(payload)
	ev := Event{Event: event, Data: string(data)}
	for c := range h.clients {
		if c.IsSupporter {
			h.send(c, ev)
		}
	}
}

func (h *Hub) SendToQuestion(event string, payload interface{}, questionUserID uint) {
	data, _ := json.Marshal(payload)
	ev := Event{Event: event, Data: string(data)}
	for c := range h.clients {
		if c.IsSupporter || c.UserID == questionUserID {
			h.send(c, ev)
		}
	}
}

func (h *Hub) SendMemo(event string, payload interface{}) {
	data, _ := json.Marshal(payload)
	ev := Event{Event: event, Data: string(data)}
	for c := range h.clients {
		if c.IsSupporter {
			h.send(c, ev)
		}
	}
}
