package channelhub

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

//Channel is the API model
//an json example would look like:
// {"id":"myChannelID", "owner":"user_id", "rights":"rwrw", "grant":{"user_id":"rw", "user_id2":"r-", "user_id3":"rw"}}
//permission can be updated by the user only if it is not a system or a dynamic channel
type Channel struct {
	//There are three types of channels that is handled over the ID
	// ----- Type ------ | ---------- example -------------- | ------ description ------
	// * static  channel | /abc                              | One channel with one static id and static config
	// * dynamic channel | /abc/:id/:param1                  | Multiple channels with dynamic id's which allows you to receive the parameters by calling IDParam("param1") and create a config for all channels with this pattern.
	// * owner   channel | /me/:Owner or /me/:Owner/:someID  | One or multiple channels with read and write rights for owner and system only.
	ID string `json:"id"`
	//idParams contains the named parameters of a dynamic channel like myChannelID/:namedParam
	//it is nil if it's not a dynamic channel or if it has no named parameters
	idParams map[string]string
	//Owner defines the id of the client that created this channel
	//it is empty if it is a system channel
	//Update Rights: @Owner only!
	Owner string `json:"owner,omitempty"`
	//Update Rights: @Owner only!
	Group string `json:"group,omitempty"`
	//Rights pattern:       	    group others
	//		                     --     --
	//default value for none system is: 	----
	//default value for system is:	    	r-r-
	//Update Rights: @Owner only!
	Rights string `json:"rights,omitempty"`
	//Grant can be modified by the owner only and is an optional field to whitelist user's directly
	//Update Rights: everyone with write rights!
	Grant map[string]string `json:"grant,omitempty"`
	//System can only be set by the ChannelHub API not from the client
	//it stands for auto subscribe, read only for clients and it is not going to be removed from the runtime
	//Update Rights: @System only!
	System bool `json:"system,omitempty"`

	//async listeners
	Created      ChannelListener `json:"-"`
	Subscribed   ChannelListener `json:"-"`
	Unsubscribed ChannelListener `json:"-"`
	Removed      ChannelListener `json:"-"`

	//sync listener
	BeforeBroadcast func(c *Channel, chMsg *ChannelHubMsg) (proceed bool) `json:"-"`

	clients      map[*client]bool `json:"-"`
	clientsCount int              `json:"-"`

	dynChanl  *dynamicChannel `json:"-"`
	clientsRW sync.RWMutex    `json:"-"`
}

var (
	dynamicParamDefRegex     = regexp.MustCompile("\\:[a-zA-Z0-9]+")
	dynamicParamRegexPattern = "([a-zA-Z0-9_\\-]+)"
	reservedParamOwner       = ":Owner"
	me                       = "me"
)

type dynamicChannel struct {
	regx           *regexp.Regexp
	idRegexPattern string
	chanl          *Channel
	paramNames     []string
	isOwnerChannel bool
}

//make sure the channel is well formed
func (c *Channel) beforeAttach() error {
	if c.ID == "" {
		return errors.New("ID must be set!")
	}
	if c.ID == me {
		return errors.New("ID cannot be 'me'!")
	}
	if c.Grant != nil && len(c.Grant) > 0 {
		fg := false
		for k, v := range c.Grant {
			fg = false
			for len(c.Rights) < 2 {
				v += "-"
				fg = true
			}
			if fg {
				c.Grant[k] = v
			}
		}
	}
	if c.System {
		if c.Rights == "" {
			c.Rights = "r-r-"
		}
	}
	for len(c.Rights) < 4 {
		c.Rights += "-"
	}

	if c.clients == nil {
		c.clients = make(map[*client]bool)
	}
	return nil
}

//Broadcast must be called by the system only to prevent from security issues or other kind of cycles.
//By the system only means no call from the client side should be able to trigger this function directly.
func (c *Channel) Broadcast(d interface{}) {
	if c.hasSubscribers() {
		cMsg := &ChannelHubMsg{Method: Method_Publish, ChannelID: c.ID, Data: d}
		if c.notifyBroadcastAndProceed(cMsg) {
			broadcastBytes, err := json.Marshal(cMsg)
			if err == nil {
				c.broadcast(cMsg.client, &broadcastBytes)
			}
		}
	}
}

func (c *Channel) tryBroadcast(cMsg *ChannelHubMsg) {
	if c.hasSubscribers() {
		if c.isWriteGrantedFor(cMsg.client) {
			if c.notifyBroadcastAndProceed(cMsg) {
				broadcastBytes, err := json.Marshal(cMsg)
				if err == nil {
					c.broadcast(cMsg.client, &broadcastBytes)
				}
			}
		}
	}
}

//IDParam makes possible to get a named parameter of the dynamic channels ID
//If it is not a dynamic channel you'll get ""
//If the dynamic channel was initiated with DynamicChannel(&Channel{ID:"channel:p1",...})
//then the runtime ID might be "channelABC-123"
//and then by calling IDParam("p1") you'll get "ABC-123"
func (c *Channel) IDParam(name string) string {
	if c.dynChanl != nil {
		return c.dynChanl.param(c, name)
	}
	return ""
}

//IsDynamic tells you if it was initiated by ChannelHub.DynamicChannel
//If it is dynamic no user can update the permissions of the channel
func (c *Channel) IsDynamic() bool {
	return c.dynChanl != nil
}

func (c *Channel) isClientAllowedToChangePermissions() bool {
	return !c.System && !c.IsDynamic()
}

func (c *Channel) subscribe(ch *ChannelHub, cMsg *ChannelHubMsg) {
	if c.isReadGrantedFor(cMsg.client) {
		c.subscribeResource(cMsg.client, func(chnl *Channel) {
			if !chnl.System {
				ch.notify <- &notifyMsg{kind: Notify_Subscribed, client: cMsg.client, channel: chnl}
				cMsg.Data = "ok"
				resp, err := json.Marshal(cMsg)
				if err == nil {
					cMsg.client.SendBytes(&resp)
					chnl.notifySubscribers(cMsg.client, &resp)
				}
			}
		})
	}
}

func (c *Channel) unsubscribe(ch *ChannelHub, cMsg *ChannelHubMsg, disconnect bool) {
	c.unsubscribeResource(cMsg.client, disconnect, func(chnl *Channel) {
		ch.notify <- &notifyMsg{kind: Notify_Unsubscribed, client: cMsg.client, channel: chnl}
		var resp []byte
		if !disconnect {
			cMsg.Data = "ok"
			var err error
			resp, err = json.Marshal(cMsg)
			if err == nil {
				cMsg.client.SendBytes(&resp)
			}
		}
		if ch.channelNotRemoved(chnl) {
			if resp == nil {
				var err error
				resp, err = json.Marshal(cMsg)
				if err == nil {
					chnl.notifySubscribers(cMsg.client, &resp)
				}
			} else {
				chnl.notifySubscribers(cMsg.client, &resp)
			}
		} else {
			ch.notify <- &notifyMsg{kind: Notify_Removed, client: cMsg.client, channel: chnl}
		}
	})
}

func (c *Channel) isReadGrantedFor(client *client) bool {
	if client == nil {
		//must be the system
		return true
	}
	//check owner
	if client.id != "" {
		if client.id == c.Owner {
			return true
		}

		if c.Grant != nil && len(c.Grant) > 0 {
			rights := c.Grant[client.id]
			if rights != "" && rights[0:1] == "r" {
				return true
			}
		}
	}
	if client.session != nil {
		//check group
		if client.session.Group != "" {
			if c.Group != "" {
				if client.session.Group == c.Group {
					//same group
					if c.Rights[0:1] == "r" {
						//has read rights
						return true
					}
				}
			}
		}
	}
	//check other
	if c.Rights[2:3] == "r" {
		//has read rights
		return true
	}
	return false
}

func (c *Channel) isWriteGrantedFor(client *client) bool {
	if client == nil {
		//must be the system
		return true
	}
	if c.System {
		//system channel but user wants to write
		return false
	}
	//check owner
	if client.id != "" {
		if client.id == c.Owner {
			return true
		}

		if c.Grant != nil && len(c.Grant) > 0 {
			rights := c.Grant[client.id]
			if rights != "" && rights[1:2] == "w" {
				return true
			}
		}
	}
	if client.session != nil {
		//check group
		if client.session.Group != "" {
			if c.Group != "" {
				if client.session.Group == c.Group {
					//same group
					if c.Rights[1:2] == "w" {
						//has write rights
						return true
					}
				}
			}
		}
	}
	//check other
	if c.Rights[3:4] == "w" {
		//has write rights
		return true
	}
	return false
}

func (c *Channel) isOwner(clnt *client) bool {
	if clnt != nil && clnt.id != "" {
		if clnt.id == c.Owner {
			return true
		}
	}
	return false
}

func (c *Channel) updatePermissions(cMsg *ChannelHubMsg) (updated bool) {
	if c.isClientAllowedToChangePermissions() {
		if c.isOwner(cMsg.client) {
			if cMsg.Channel.Owner != "" {
				c.Owner = cMsg.Channel.Owner
			}
			if cMsg.Channel.Group != "" {
				c.Group = cMsg.Channel.Group
			}
			if cMsg.Channel.Rights != "" {
				c.Rights = cMsg.Channel.Rights
			}

			if cMsg.Channel.Grant != nil && len(cMsg.Channel.Grant) > 0 {
				if c.Grant != nil {
					for k, item := range cMsg.Channel.Grant {
						c.Grant[k] = item
					}
				} else {
					c.Grant = cMsg.Channel.Grant
				}
			}
			c.beforeAttach()
			updated = true
		} else {
			if c.isWriteGrantedFor(cMsg.client) {
				if cMsg.Channel.Grant != nil && len(cMsg.Channel.Grant) > 0 {
					if c.Grant != nil {
						for k, item := range cMsg.Channel.Grant {
							c.Grant[k] = item
						}
					} else {
						c.Grant = cMsg.Channel.Grant
					}
					c.beforeAttach()
					updated = true
				}
			}
		}
	}
	return
}

func (c *Channel) hasSubscribers() bool {
	return c.clientsCount > 0
}

func (c *Channel) subscribeResource(client *client, notify func(chanl *Channel)) {
	alreadySubscribed := c.clients[client]
	if !alreadySubscribed {
		c.clientsRW.Lock()
		c.clients[client] = true
		c.clientsCount = len(c.clients)
		c.clientsRW.Unlock()

		client.channels[c] = true

		if notify != nil {
			notify(c)
		}
	}
}

func (c *Channel) unsubscribeResource(client *client, disconnect bool, notify func(chanl *Channel)) {
	if disconnect || !c.System {
		notUnsubscribed := c.clients[client]
		if notUnsubscribed {
			c.clientsRW.Lock()
			delete(c.clients, client)
			c.clientsCount = len(c.clients)
			c.clientsRW.Unlock()

			delete(client.channels, c)

			if !c.System && notify != nil {
				notify(c)
			}
		}
	}

}

func (c *Channel) shouldBeRemoved() bool {
	return !c.System && c.clientsCount == 0
}

func (c *Channel) notifySubscribers(client *client, b *[]byte) {
	c.broadcast(client, b)
}

func (c *Channel) broadcast(sender *client, b *[]byte) {
	c.clientsRW.RLock()
	for clnt := range c.clients {
		if sender != clnt {
			clnt.send <- b
		}
	}
	c.clientsRW.RUnlock()
}

func (c *Channel) hasNamedParameter() bool {
	return len(dynamicParamDefRegex.FindAllString(c.ID, -1)) > 0
}

func (c *Channel) makeDynamicChannel() (*dynamicChannel, error) {
	if c.ID == "" {
		return nil, errors.New("you must provide a channel ID in form of text like 'myChannelID' or with named parameters like 'myChannelID/:channelParam1'!")
	}
	//reg, err := regexp.Compile(c.ID)
	//if err != nil {
	//	return err
	//}
	dynChanl := &dynamicChannel{chanl: c}
	dynChanl.isOwnerChannel = strings.Contains(c.ID, reservedParamOwner)
	dynChanl.idRegexPattern = c.ID

	params := dynamicParamDefRegex.FindAllString(c.ID, -1)
	lp := len(params)
	if lp > 0 {
		dynChanl.paramNames = make([]string, lp)
		for i, item := range params {
			dynChanl.idRegexPattern = strings.Replace(dynChanl.idRegexPattern, item, dynamicParamRegexPattern, -1)
			if len(item) > 0 {
				dynChanl.paramNames[i] = item[1:]
			}
		}

		r2, err := regexp.Compile(dynChanl.idRegexPattern)
		if err != nil {
			return nil, err
		}
		dynChanl.regx = r2
	} else {
		r2, err := regexp.Compile(dynChanl.idRegexPattern)
		if err != nil {
			return nil, err
		}
		dynChanl.regx = r2
	}
	fmt.Println(c.ID, dynChanl.paramNames)
	return dynChanl, nil
}

func (dc *dynamicChannel) createChannel(cMsg *ChannelHubMsg) *Channel {
	return &Channel{
		ID:              cMsg.ChannelID,
		Owner:           dc.chanl.Owner,
		Group:           dc.chanl.Group,
		Rights:          dc.chanl.Rights,
		Grant:           dc.chanl.Grant,
		System:          dc.chanl.System,
		Created:         dc.chanl.Created,
		Subscribed:      dc.chanl.Subscribed,
		Unsubscribed:    dc.chanl.Unsubscribed,
		Removed:         dc.chanl.Removed,
		BeforeBroadcast: dc.chanl.BeforeBroadcast,
		dynChanl:        dc,
	}
}

func (dc *dynamicChannel) param(chanl *Channel, name string) string {
	if chanl.idParams == nil {
		if len(dc.paramNames) > 0 {
			//runtime
			pmap := make(map[string]string)
			ps := dc.regx.FindAllStringSubmatch(chanl.ID, -1)
			if len(ps) > 0 {
				a := ps[0]
				la := len(a)
				if la > 1 {
					for i := 1; i < la; i++ {
						pmap[dc.paramNames[i-1]] = a[i]
					}
				}
			}
			chanl.idParams = pmap
			return chanl.idParams[name]
		}
	} else {
		return chanl.idParams[name]
	}
	return ""
}

func (dc *dynamicChannel) isDynamicChannel(cMsg *ChannelHubMsg) bool {
	machtes := dc.regx.MatchString(cMsg.ChannelID)
	if machtes && dc.isOwnerChannel {
		if cMsg.client != nil {
			return (&Channel{ID: cMsg.ChannelID}).IDParam("Owner") == cMsg.client.id
		}
		return false
	}
	return machtes
}
