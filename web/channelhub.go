package channelhub

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		//CheckOrigin: func(r *http.Request) bool { return true },
	}

	ErrHubNotRunning = errors.New("hub not running yet!")
)

const (
	Method_Publish     = "pub"
	Method_Subscribe   = "sub"
	Method_Unsubscribe = "unsub"
	Method_Update      = "update"
	Method_Data        = "data"
	//Method_System is needed to let the client know what the system channels are
	//this way the client knows to respond immediately with an error in case someone tries to push something
	Method_System = "sys"

	Notify_Created      = 1
	Notify_Subscribed   = 2
	Notify_Unsubscribed = 3
	Notify_Removed      = 4
)

type (
	ChannelListener func(channel *Channel, client *Client)

	ChannelHub struct {
		ChannelFind         func(cMsg *ChannelHubMsg) (chnl *Channel, create bool)
		ChannelDataFind     func(cMsg *ChannelHubMsg)
		ChannelCreated      ChannelListener
		ChannelSubscribed   ChannelListener
		ChannelUnsubscribed ChannelListener
		ChannelRemoved      ChannelListener
		//ClientCreateChannelLimit defines the limit of channel creation from the client side.
		//if the limit is reached no channels that don't exist on runtime or cannot be found by ChannelFind are created anymore.
		//default is 10
		ClientCreateChannelLimit int
		//default is 140
		ChannelIDLengthLimit int

		//managed by hubThread
		sessions map[string]*hSession

		registerClient   chan *regMsg
		unregisterClient chan *client
		sessionsCount    int
		hubThreadRunning bool

		channels        map[string]*Channel
		systemChannels  map[*Channel]bool
		dynamicChannels map[string]*dynamicChannel
		input           chan *ChannelHubMsg
		notify          chan *notifyMsg

		sessionsRW sync.RWMutex

		closed bool
	}

	ChannelHubMsg struct {
		Method    string      `json:"m,omitempty"`
		ChannelID string      `json:"cid,omitempty"`
		Channel   *Channel    `json:"c,omitempty"`
		RequestID int         `json:"rid,omitempty"`
		Data      interface{} `json:"d,omitempty"`
		//ClientID should be the same as Owner
		ClientID string  `json:"u,omitempty"`
		client   *client `json:"-"`
	}

	regMsg struct {
		sessionID   string
		clientID    string
		clientGroup string
		client      *client
	}

	hSession struct {
		ID              string
		Owner           string
		Group           string
		Clients         map[*client]bool
		closed          bool
		ClientsCount    int
		hub             *ChannelHub
		createdChannels int

		clientsRW sync.RWMutex
	}

	notifyMsg struct {
		kind    int8
		client  *client
		channel *Channel
	}
)

func (ch *ChannelHub) Run(startUpChannels ...*Channel) error {
	if ch.hubThreadRunning {
		return errors.New("already running! ")
	}
	if ch.ClientCreateChannelLimit == 0 {
		ch.ClientCreateChannelLimit = 10
	}
	if ch.ChannelIDLengthLimit == 0 {
		ch.ChannelIDLengthLimit = 140
	}
	ch.sessions = make(map[string]*hSession)
	ch.registerClient = make(chan *regMsg, 200)
	ch.unregisterClient = make(chan *client, 200)
	ch.sessionsCount = 0

	ch.input = make(chan *ChannelHubMsg, 200)
	ch.notify = make(chan *notifyMsg, 200)

	if ch.channels == nil {
		ch.channels = make(map[string]*Channel)
	} else {
		for _, item := range ch.channels {
			item.clients = make(map[*client]bool)
			item.clientsCount = 0
		}
	}

	if ch.systemChannels == nil {
		ch.systemChannels = make(map[*Channel]bool)
	}

	if startUpChannels != nil {
		l := len(startUpChannels)
		if l > 0 {
			var chanl *Channel
			var err error
			for i := 0; i < l; i++ {
				chanl = startUpChannels[i]
				if err = chanl.beforeAttach(); err == nil {
					ch.attach(chanl)
				} else {
					fmt.Errorf("Validation error with Channel %s : %s", chanl.ID, err.Error())
				}
			}
		}
	}
	go ch.hubThread()
	go ch.notifyThread()
	return nil
}

//Put must be called before the Run function to prevent from multi thread issues and unnecessary locks or golang channels
func (ch *ChannelHub) Put(chanls ...*Channel) error {
	if ch.hubThreadRunning {
		return errors.New("you must call this function before you call Run!")
	}
	if len(chanls) == 0 {
		return errors.New("you must provide at least one channel pointer!")
	}
	for _, chanl := range chanls {
		if chanl.hasNamedParameter() { // dynamic channel
			alreadyExists := ch.dynamicChannels[chanl.ID]
			if alreadyExists == nil {
				dynChanl, err := chanl.makeDynamicChannel()
				if err != nil {
					return err
				}
				if ch.dynamicChannels == nil {
					ch.dynamicChannels = make(map[string]*dynamicChannel)
				}
				ch.dynamicChannels[chanl.ID] = dynChanl
			}
		} else { // static channel
			alreadyExists := ch.channels[chanl.ID]
			if alreadyExists == nil {
				err := chanl.beforeAttach()
				if err != nil {
					return err
				}
				ch.attach(chanl)
			}
		}
	}
	return nil
}

//Broadcast must be called by the system only to prevent from security issues or other kind of cycles.
//By the system only means no call from the client side should be able to trigger this function directly.
func (ch *ChannelHub) Broadcast(chanlID string, d interface{}) error {
	if !ch.hubThreadRunning || ch.input == nil {
		return ErrHubNotRunning
	}
	ch.input <- &ChannelHubMsg{Method: Method_Publish, ChannelID: chanlID, Data: d}
	return nil
}

//BroadcastToUser must be called by the system only to prevent from security issues or other kind of cycles.
//By the system only means no call from the client side should be able to trigger this function directly.
func (ch *ChannelHub) BroadcastToUser(usrID string, d interface{}) error {
	if !ch.hubThreadRunning {
		return ErrHubNotRunning
	}
	cMsg := &ChannelHubMsg{Method: Method_Publish, ChannelID: "me", Data: d}
	ch.sessionsRW.RLock()
	s := ch.sessions[usrID]
	ch.sessionsRW.RUnlock()
	if s != nil {
		broadcastBytes, err := json.Marshal(cMsg)
		if err == nil {
			s.sendToAllClients(&broadcastBytes)
		}
		return nil
	}
	return nil
}

func (ch *ChannelHub) Close() {
	if ch.closed {
		return
	}
	ch.closed = true
	if ch.registerClient != nil {
		close(ch.input)
		close(ch.registerClient)
		close(ch.unregisterClient)
		for {
			if !ch.hubThreadRunning {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		close(ch.notify)
	}
}

func (ch *ChannelHub) NewClient(w http.ResponseWriter, r *http.Request, sessionID, userID, userGroup string) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("NewClient: ", sessionID)
	client := &client{
		id:        userID,
		sessionID: sessionID,
		hub:       ch,
		conn:      conn,
		send:      make(chan *[]byte, 20),
	}
	hubm := &regMsg{sessionID: sessionID, clientID: userID, clientGroup: userGroup, client: client}
	if ch.hubThreadRunning && ch.registerClient != nil {
		ch.registerClient <- hubm
	}
	return nil
}

func (c *Channel) notifyBroadcastAndProceed(chMsg *ChannelHubMsg) bool {
	if c.BeforeBroadcast != nil {
		return c.BeforeBroadcast(c, chMsg)
	}
	return true
}

func (ch *ChannelHub) hubThread() {
	ch.hubThreadRunning = true
	defer func() {
		ch.hubThreadRunning = false
		fmt.Println("Shutting down hubThread!")
		ch.sessionsRW.Lock()
		for _, item := range ch.sessions {
			item.Close()
		}
		ch.sessions = nil
		ch.sessionsRW.Unlock()
	}()
	for {
		select {
		case cMsg, oK := <-ch.input:
			if !oK {
				return
			}
			switch cMsg.Method {
			case Method_Publish:
				chanl := ch.channels[cMsg.ChannelID]
				if chanl != nil {
					chanl.tryBroadcast(cMsg)
				}
				break
			case Method_Subscribe:
				if cMsg.clientNotNil() {
					chanl := ch.channels[cMsg.ChannelID]
					if chanl == nil {
						chanl = ch.findChannelAndAttach(cMsg)
					}
					if chanl != nil {
						chanl.subscribe(ch, cMsg)
					}
				}
				break
			case Method_Update:
				if cMsg.Channel != nil && cMsg.clientNotNil() {
					chanl := ch.channels[cMsg.ChannelID]
					if chanl != nil {
						updated := chanl.updatePermissions(cMsg)
						if updated {
							b, err := json.Marshal(cMsg)
							if err == nil {
								cMsg.client.SendBytes(&b)
								chanl.notifySubscribers(cMsg.client, &b)
							}
						}
					}
				}
				break
			case Method_Unsubscribe:
				if cMsg.clientNotNil() {
					chanl := ch.channels[cMsg.ChannelID]
					if chanl != nil && !chanl.System {
						chanl.unsubscribe(ch, cMsg, false)
					}
				}
				break
			}
		case hubm, oK := <-ch.registerClient:
			if !oK {
				return
			}
			ch.sessionsRW.RLock()
			s := ch.sessions[hubm.clientID]
			ch.sessionsRW.RUnlock()
			hubm.client.channels = make(map[*Channel]bool)
			if s == nil {
				s = &hSession{
					ID:           hubm.sessionID,
					Owner:        hubm.clientID,
					Group:        hubm.clientGroup,
					hub:          ch,
					Clients:      map[*client]bool{hubm.client: true},
					ClientsCount: 1,
					closed:       false,
				}
				hubm.client.session = s
				ch.sessionsRW.Lock()
				ch.sessions[hubm.clientID] = s
				ch.sessionsCount = len(ch.sessions)
				ch.sessionsRW.Unlock()
			} else {
				hubm.client.session = s
				s.clientsRW.Lock()
				s.Clients[hubm.client] = true
				s.ClientsCount = len(s.Clients)
				s.clientsRW.Unlock()
			}
			hubm.client.StartIO()
			// silent auto subscribe system channels
			cMsg := &ChannelHubMsg{Method: Method_System}
			syschanlLength := len(ch.systemChannels)
			syschanls := make([]string, syschanlLength)
			i := 0
			for canl := range ch.systemChannels {
				canl.subscribeResource(hubm.client, nil)
				syschanls[i] = canl.ID
				i++
			}
			cMsg.Data = syschanls
			hubm.client.Send(cMsg)
		case client, oK := <-ch.unregisterClient:
			if !oK {
				return
			}
			if client != nil {
				client.disconnect()
			}
		}
	}
}

func (ch *ChannelHub) findChannelAndAttach(cMsg *ChannelHubMsg) (chanl *Channel) {
	var isDynamic bool
	chanl, isDynamic = ch.isDynamicChannel(cMsg)
	if isDynamic {
		if chanl.beforeAttach() == nil {
			ch.notify <- &notifyMsg{kind: Notify_Created, client: cMsg.client, channel: chanl}
			ch.attach(chanl)
			return
		}
	} else {
		if ch.ChannelFind != nil {
			var allowedToCreate bool
			chanl, allowedToCreate = ch.ChannelFind(cMsg)
			if chanl != nil {
				if chanl.beforeAttach() == nil {
					ch.notify <- &notifyMsg{kind: Notify_Created, client: cMsg.client, channel: chanl}
					ch.attach(chanl)
					return
				}
			} else if allowedToCreate {
				if cMsg.client.ChannelLimitNotReached() {
					chanl = cMsg.createClientChannelAndAttach()
					return
				}
			}
		} else {
			if cMsg.client.ChannelLimitNotReached() {
				chanl = cMsg.createClientChannelAndAttach()
				return
			}
		}
	}
	chanl = nil
	return
}

func (ch *ChannelHub) attach(chanl *Channel) {
	ch.channels[chanl.ID] = chanl
	if chanl.System {
		ch.systemChannels[chanl] = true
	}
}

func (ch *ChannelHub) channelNotRemoved(chanl *Channel) bool {
	if chanl.shouldBeRemoved() {
		delete(ch.channels, chanl.ID)
		return false
	}
	return true
}

func (cMsg *ChannelHubMsg) createClientChannelAndAttach() *Channel {
	if cMsg.Channel == nil {
		cMsg.Channel = &Channel{
			ID:    cMsg.ChannelID,
			Owner: cMsg.client.id,
			Group: cMsg.client.session.Group,
		}
	}
	cMsg.Channel.System = false // ensure it is false
	if cMsg.Channel.beforeAttach() == nil {
		cMsg.client.hub.notify <- &notifyMsg{kind: Notify_Created, client: cMsg.client, channel: cMsg.Channel}
		cMsg.client.session.createdChannels++
		cMsg.client.hub.attach(cMsg.Channel)
		return cMsg.Channel
	}
	return nil
}

func (cMsg *ChannelHubMsg) clientNotNil() bool {
	return cMsg.client != nil
}

func (cMsg *ChannelHubMsg) session() *hSession {
	if cMsg.client != nil {
		return cMsg.client.session
	}
	return nil
}

func (ch *ChannelHub) isDynamicChannel(cMsg *ChannelHubMsg) (*Channel, bool) {
	if ch.dynamicChannels != nil {
		for _, item := range ch.dynamicChannels {
			if item.isDynamicChannel(cMsg) {
				return item.createChannel(cMsg), true
			}
		}
	}
	return nil, false
}

func (s *hSession) channelLimitReached() bool {
	return s.createdChannels+1 > s.hub.ClientCreateChannelLimit
}

func (ch *ChannelHub) notifyThread() {
	defer func() {
		fmt.Println("Shutting down notifyThread!")
	}()
	for {
		select {
		case nMsg, oK := <-ch.notify:
			if !oK {
				return
			}
			switch nMsg.kind {
			case Notify_Created:
				nMsg.notifyGlobal(ch.ChannelCreated)
				break
			case Notify_Subscribed:
				nMsg.notifyGlobal(ch.ChannelSubscribed)
				break
			case Notify_Unsubscribed:
				nMsg.notifyGlobal(ch.ChannelUnsubscribed)
				break
			case Notify_Removed:
				nMsg.notifyGlobal(ch.ChannelRemoved)
				break
			}
		}
	}
}

func (nMsg *notifyMsg) notifyGlobal(fn ChannelListener) {
	if fn != nil {
		var cl Client
		if nMsg.client != nil {
			cl = nMsg.client.makePublic()
		}
		if nMsg.channel != nil {
			//channel listener
			switch nMsg.kind {
			case Notify_Created:
				nMsg.notifyChannel(nMsg.channel.Created, &cl)
				break
			case Notify_Subscribed:
				nMsg.notifyChannel(nMsg.channel.Subscribed, &cl)
				break
			case Notify_Unsubscribed:
				nMsg.notifyChannel(nMsg.channel.Unsubscribed, &cl)
				break
			case Notify_Removed:
				nMsg.notifyChannel(nMsg.channel.Removed, &cl)
				break
			}
		}
		//global listener
		fn(nMsg.channel, &cl)
	}
}

func (nMsg *notifyMsg) notifyChannel(fn ChannelListener, cl *Client) {
	if fn != nil {
		fn(nMsg.channel, cl)
	}
}

func (s *hSession) sendToAllClients(b *[]byte) {
	s.clientsRW.RLock()
	defer s.clientsRW.RUnlock()
	for client := range s.Clients {
		client.SendBytes(b)
	}
}

func (s *hSession) disconnect(c *client) (empty bool) {
	s.clientsRW.Lock()
	delete(s.Clients, c)
	s.ClientsCount = len(s.Clients)
	s.clientsRW.Unlock()
	empty = s.ClientsCount == 0
	return
}

func (s *hSession) Close() {
	s.clientsRW.Lock()
	defer s.clientsRW.Unlock()
	for citem := range s.Clients {
		citem.Close()
	}
}
