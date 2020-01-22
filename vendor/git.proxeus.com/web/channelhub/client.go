package channelhub

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	//maxMessageSize = 10512
)

type (
	Client struct {
		client *client `json:"-"`
	}

	client struct {
		id        string //Owner/UserID
		sessionID string
		session   *hSession
		hub       *ChannelHub
		conn      *websocket.Conn
		closed    bool
		send      chan *[]byte
		channels  map[*Channel]bool
	}
)

func (c *Client) ID() string {
	return c.client.id
}

func (c *Client) SessionID() string {
	return c.client.sessionID
}

func (c *Client) Group() string {
	if c.client.session != nil {
		return c.client.session.Group
	}
	return ""
}

func (c *Client) LocalAddr() net.Addr {
	if c.client.conn != nil {
		return c.client.conn.LocalAddr()
	}
	return nil
}

func (c *Client) RemoteAddr() net.Addr {
	if c.client.conn != nil {
		return c.client.conn.RemoteAddr()
	}
	return nil
}

func (c *client) makePublic() Client {
	return Client{client: c}
}

func (c *client) StartIO() {
	go c.writeThread()
	go c.readThread()
}

func (c *client) Send(d interface{}) {
	resp, err := json.Marshal(d)
	if err == nil {
		if !c.closed && c.send != nil {
			c.send <- &resp
		}
	}
}

func (c *client) SendBytes(b *[]byte) {
	if !c.closed && c.send != nil {
		c.send <- b
	}
}

func (c *client) Close() {
	if !c.closed {
		c.closed = true
		log.Println("Client->Close")
		c.conn.Close()
		close(c.send)
	}
}

func (c *client) ChannelLimitNotReached() bool {
	return c.session != nil && !c.session.channelLimitReached()
}

func (c *client) readThread() {
	log.Println("readThread start", c.id)
	defer func() {
		c.conn.Close()
		if c.hub.hubThreadRunning && c.hub.unregisterClient != nil {
			c.hub.unregisterClient <- c
		}
		log.Println("readThread close", c.id)
	}()
	//c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	var err error
	var message []byte
	for {
		_, message, err = c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		cMsg := ChannelHubMsg{}
		err = json.Unmarshal(message, &cMsg)
		if err != nil {
			log.Println(err)
			continue
		}
		if cMsg.Method == "" {
			continue
		}
		if cMsg.ChannelID == "" && cMsg.Channel != nil && cMsg.Channel.ID != "" {
			cMsg.ChannelID = cMsg.Channel.ID
		}
		if len(cMsg.ChannelID) > c.hub.ChannelIDLengthLimit {
			continue
		}
		if cMsg.ChannelID == "" || cMsg.ChannelID == me {
			continue
		}
		cMsg.ClientID = c.id
		cMsg.client = c
		if cMsg.Method == Method_Data {
			c.hub.ChannelDataFind(&cMsg)
			c.Send(cMsg)
		} else {
			c.hub.input <- &cMsg
		}
	}
}

func (c *client) writeThread() {
	log.Println("writeThread start", c.id)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		log.Println("writeThread close", c.id)
	}()
	var jsonMessage *[]byte
	var ok bool
	var err error

	for {
		select {
		case jsonMessage, ok = <-c.send:
			if !ok {
				//channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			err = c.conn.WriteMessage(websocket.TextMessage, *jsonMessage)
			if err != nil {
				return
			}
			jsonMessage = nil
			//err = c.conn.WriteJSON(jsonMessage)
			//if err != nil {
			//	return
			//}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *client) disconnect() {
	if c.session != nil {
		if c.session.disconnect(c) {
			c.hub.sessionsRW.Lock()
			delete(c.hub.sessions, c.id)
			c.hub.sessionsCount = len(c.hub.sessions)
			c.hub.sessionsRW.Unlock()
		}
	}
	if len(c.channels) > 0 {
		var cMsg *ChannelHubMsg
		for item := range c.channels {
			cMsg = &ChannelHubMsg{
				Method:    Method_Unsubscribe,
				ChannelID: item.ID,
				ClientID:  c.id,
				client:    c,
			}
			item.unsubscribe(c.hub, cMsg, true)
		}
	}
	c.session = nil
	c.hub = nil
	c.Close()
}
