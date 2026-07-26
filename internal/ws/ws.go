package ws

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

const magicGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type Client struct {
	conn io.ReadWriteCloser
}

type Hub struct {
	mu    sync.Mutex
	rooms map[string]map[*Client]bool
}

var GlobalHub = &Hub{
	rooms: make(map[string]map[*Client]bool),
}

func (h *Hub) AddClient(room string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][c] = true
}

func (h *Hub) RemoveClient(room string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] != nil {
		delete(h.rooms[room], c)
		if err := c.conn.Close(); err != nil {
			log.Printf("Error closing websocket connection: %v", err)
		}
	}
}

func (h *Hub) Broadcast(room string, msg string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.rooms[room]
	for c := range clients {
		go func(cli *Client) {
			if err := sendTextFrame(cli.conn, msg); err != nil {
				log.Printf("Error sending text frame: %v", err)
			}
		}(c)
	}
}

// ServeWS performs the HTTP WebSocket handshake using http.Hijacker (Go stdlib).
func ServeWS(w http.ResponseWriter, r *http.Request, room string) {
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "Not a websocket request", http.StatusBadRequest)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Websocket hijacking not supported", http.StatusInternalServerError)
		return
	}

	conn, bufrw, err := hijacker.Hijack()
	if err != nil {
		return
	}

	// Compute Sec-WebSocket-Accept key
	sha := sha1.New()
	sha.Write([]byte(key + magicGUID))
	acceptKey := base64.StdEncoding.EncodeToString(sha.Sum(nil))

	// Send 101 Switching Protocols header
	res := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
	if _, err := bufrw.WriteString(res); err != nil {
		log.Printf("Error writing to websocket: %v", err)
	}
	if err := bufrw.Flush(); err != nil {
		log.Printf("Error flushing websocket: %v", err)
	}

	client := &Client{conn: conn}
	GlobalHub.AddClient(room, client)

	// Keep alive & read loop until client disconnects
	go func() {
		defer GlobalHub.RemoveClient(room, client)
		buf := make([]byte, 1024)
		for {
			_, err := bufrw.Read(buf)
			if err != nil {
				break
			}
		}
	}()
}

func sendTextFrame(w io.Writer, msg string) error {
	payload := []byte(msg)
	length := len(payload)

	var frame []byte
	frame = append(frame, 0x81) // FIN bit + Text Opcode (0x1)

	if length <= 125 {
		frame = append(frame, byte(length))
	} else if length <= 65535 {
		frame = append(frame, 126, byte(length>>8), byte(length&0xFF))
	} else {
		return fmt.Errorf("message payload too long")
	}

	frame = append(frame, payload...)
	_, err := w.Write(frame)
	return err
}
