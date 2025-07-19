// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/oapi-codegen/nullable"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	channels map[string]*Channel
	mu       sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		channels: make(map[string]*Channel),
	}
}

func (h *Hub) GetOrCreateChannel(name string) *Channel {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch, ok := h.channels[name]
	if !ok {
		ch = &Channel{
			clients: make(map[*Client]bool),
			hub:     h,
			name:    name,
		}
		h.channels[name] = ch
	}
	return ch
}

func (h *Hub) GetChannelNames() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var names []string = make([]string, len(h.channels))
	i := 0
	for name := range h.channels {
		names[i] = name
		i++
	}
	return names
}

func (h *Hub) DeleteChannel(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.channels, name)
}

type Channel struct {
	mainClient *Client
	clients    map[*Client]bool // includes mainClient
	mu         sync.RWMutex
	hub        *Hub
	name       string
}

func (ch *Channel) Register(client *Client) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.clients[client] = true
	if client.IsMain {
		ch.mainClient = client
	}
}

func (ch *Channel) Unregister(client *Client) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if _, ok := ch.clients[client]; ok {
		delete(ch.clients, client)
		close(client.Send)
		if client == ch.mainClient {
			ch.mainClient = nil
		}
		if len(ch.clients) == 0 {
			ch.hub.DeleteChannel(ch.name)
		}
	}
}

func (ch *Channel) RouteMessage(sender *Client, msg []byte) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	var incoming Message
	if err := json.Unmarshal(msg, &incoming); err != nil {
		log.Printf("Malformed request from channel '%s': %v | Raw: %s", sender.ChannelName, err, string(msg))
		return
	}

	switch incoming.Op {
	case "bind":
		if sender.IsMain {
			var bindContent BindMessage
			if err := json.Unmarshal(incoming.Content, &bindContent); err != nil {
				log.Printf("Malformed bind request from channel '%s': %v", sender.ChannelName, err)
				return
			}
			response := createBindResponse(sender, incoming.SenderId, bindContent)
			responseBytes, _ := json.Marshal(response)
			sender.Send <- responseBytes
		}

	case "deviceStatus", "brewStatus", "brewingComplete", "getDeviceStatus":
		if sender.IsMain {
			// Main client sends to all other clients
			for client := range ch.clients {
				if client != sender {
					select {
					case client.Send <- msg:
					default:
						close(client.Send)
						delete(ch.clients, client)
					}
				}
			}
		} else {
			// Non-main client sends only to main client
			if ch.mainClient != nil {
				select {
				case ch.mainClient.Send <- msg:
				default:
					close(ch.mainClient.Send)
					delete(ch.clients, ch.mainClient)
				}
			}
		}
	case "messageCountRequest":
		{
		}
	default:
		log.Printf("Unhandled op '%s' from channel '%s': %s", incoming.Op, sender.ChannelName, string(msg))
	}
}

type Message struct {
	Op       string          `json:"op"`
	SenderId string          `json:"senderId"`
	Content  json.RawMessage `json:"content"`
}

type BindMessage struct {
	SerialNumber string `json:"serialNumber"`
	Code         string `json:"code"`
}

type BindResponse struct {
	BindingId  int                       `json:"bindingId"`
	DeviceId   string                    `json:"deviceId"`
	UserId     nullable.Nullable[string] `json:"userId"`
	State      string                    `json:"state"`
	ChangeDate int64                     `json:"changeDate"`
	Code       string                    `json:"code"`
	Name       nullable.Nullable[string] `json:"name"`
	User       nullable.Nullable[string] `json:"user"`
}

func createBindResponse(sender *Client, senderId string, jc BindMessage) map[string]interface{} {
	return map[string]interface{}{
		"op":               "bindingStatus",
		"receivedId":       "device://" + senderId,
		"senderId":         "broker://device",
		"clientId":         nil,
		"mime":             "application/json",
		"creationDate":     time.Now().Unix(),
		"dispatchDate":     nil,
		"changeDate":       nil,
		"timeout":          1800000,
		"useCaseId":        nil,
		"redelivered":      0,
		"durable":          true,
		"async":            false,
		"messageStatus":    "Pending",
		"destinationTopic": nil,
		"scheduledDelay":   0,
		"softwareVersion":  nil,
		"content": BindResponse{
			BindingId:  6805, // This seems to be a magic number
			DeviceId:   senderId,
			UserId:     nullable.Nullable[string]{},
			State:      "BOUND",
			ChangeDate: time.Now().Unix(),
			Code:       jc.Code,
			Name:       nullable.Nullable[string]{},
			User:       nullable.Nullable[string]{},
		},
	}
}
