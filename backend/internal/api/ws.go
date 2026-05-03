package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/cryptotrade/app/internal/aggregator"
	"github.com/cryptotrade/app/internal/exchange"
	"github.com/cryptotrade/app/internal/hub"
)

type WSHandler struct {
	hub       *hub.Hub
	agg       *aggregator.Aggregator
	upgrader  *websocket.Upgrader
	jwtSecret string
}

func NewWSHandler(h *hub.Hub, agg *aggregator.Aggregator, upgrader *websocket.Upgrader, secret string) *WSHandler {
	return &WSHandler{hub: h, agg: agg, upgrader: upgrader, jwtSecret: secret}
}

type snapshotMsg struct {
	Type          string                         `json:"type"`
	Ticks         map[string]exchange.PriceTick  `json:"ticks"`
	Opportunities []aggregator.SpreadOpportunity `json:"opportunities"`
}

func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	claims, err := validateJWT(tokenStr, h.jwtSecret)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &hub.Client{
		UserID: claims.UserID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}
	h.hub.Register(client)
	defer func() {
		h.hub.Unregister(client.UserID)
		conn.Close()
	}()

	snap := snapshotMsg{
		Type:          "snapshot",
		Ticks:         h.agg.GetTicks(),
		Opportunities: h.agg.ComputeOpportunities(),
	}
	if data, err := json.Marshal(snap); err == nil {
		client.Send <- data
	}

	go h.writePump(client)
	h.readPump(client)
}

func (h *WSHandler) writePump(c *hub.Client) {
	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()
	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, msg)
		case <-ping.C:
			c.Conn.WriteMessage(websocket.PingMessage, nil)
		}
	}
}

func (h *WSHandler) readPump(c *hub.Client) {
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		if _, _, err := c.Conn.ReadMessage(); err != nil {
			break
		}
	}
}
