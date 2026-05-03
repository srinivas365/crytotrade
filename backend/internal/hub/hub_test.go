package hub

import (
	"testing"
)

func TestHub_RegisterUnregister(t *testing.T) {
	h := New()
	c := &Client{UserID: "user1", Send: make(chan []byte, 1)}
	h.Register(c)
	if users := h.ConnectedUsers(); len(users) != 1 {
		t.Fatalf("expected 1 connected user, got %d", len(users))
	}
	h.Unregister("user1")
	if users := h.ConnectedUsers(); len(users) != 0 {
		t.Fatalf("expected 0 connected users, got %d", len(users))
	}
}

func TestHub_SendToUser(t *testing.T) {
	h := New()
	c := &Client{UserID: "user1", Send: make(chan []byte, 1)}
	h.Register(c)
	h.SendToUser("user1", []byte("hello"))
	msg := <-c.Send
	if string(msg) != "hello" {
		t.Fatalf("got %s", string(msg))
	}
}

func TestHub_Broadcast(t *testing.T) {
	h := New()
	c1 := &Client{UserID: "u1", Send: make(chan []byte, 1)}
	c2 := &Client{UserID: "u2", Send: make(chan []byte, 1)}
	h.Register(c1)
	h.Register(c2)
	h.Broadcast([]byte("ping"))
	if string(<-c1.Send) != "ping" || string(<-c2.Send) != "ping" {
		t.Fatal("broadcast failed")
	}
}
