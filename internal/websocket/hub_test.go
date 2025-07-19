package websocket

import (
	"encoding/json"
	"testing"
	"time"
)

func TestChannelMainAndSecondaryClients(t *testing.T) {
	hub := NewHub()
	channelName := "chan1"
	ch := hub.GetOrCreateChannel(channelName)

	mainClient := &Client{
		Hub:         hub,
		Send:        make(chan []byte, 2),
		ChannelName: channelName,
		IsMain:      true,
	}
	secondary := &Client{
		Hub:         hub,
		Send:        make(chan []byte, 2),
		ChannelName: channelName,
		IsMain:      false,
	}

	ch.Register(mainClient)
	ch.Register(secondary)

	if len(ch.clients) != 2 {
		t.Fatalf("expected 2 clients in channel, got %d", len(ch.clients))
	}
	if ch.mainClient != mainClient {
		t.Fatalf("main client not set correctly")
	}

	// Simulate secondary sending a message (should be routed to main)
	msg := []byte("{\"op\": \"deviceStatus\", \"senderId\": \"device://chan1\", \"content\": {\"channel\": \"chan1\", \"text\": \"hello main\"}}")
	ch.RouteMessage(secondary, msg)

	// Check main client receives the message
	select {
	case got := <-mainClient.Send:
		if string(got) != string(msg) {
			t.Errorf("main client got wrong message: got %q, want %q", got, msg)
		}
	case <-time.After(time.Second):
		t.Fatal("main client did not receive message from secondary")
	}

	// Simulate main sending a message (should be routed to all others)
	msg2 := []byte("{\"op\": \"deviceStatus\", \"senderId\": \"device://chan1\", \"content\": {\"channel\": \"chan1\", \"text\": \"hello secondary\"}}")
	ch.RouteMessage(mainClient, msg2)

	// Check secondary client receives the message
	select {
	case got := <-secondary.Send:
		if string(got) != string(msg2) {
			t.Errorf("secondary client got wrong message: got %q, want %q", got, msg2)
		}
	case <-time.After(time.Second):
		t.Fatal("secondary client did not receive message from main")
	}
}

func TestBindOperationResponse(t *testing.T) {
	hub := NewHub()
	channelName := "bindchan"
	ch := hub.GetOrCreateChannel(channelName)

	client := &Client{
		Hub:         hub,
		Send:        make(chan []byte, 2),
		ChannelName: channelName,
		IsMain:      true,
	}

	ch.Register(client)

	if len(ch.clients) != 1 {
		t.Fatal("client not registered correctly")
	}

	// Prepare bind message
	bindMsg := Message{
		Op:       "bind",
		SenderId: "test-device",
	}
	bindContent := BindMessage{
		SerialNumber: "SN1234",
		Code:         "BINDCODE",
	}
	contentBytes, err := json.Marshal(bindContent)
	if err != nil {
		t.Fatalf("failed to marshal bind content: %v", err)
	}
	bindMsg.Content = contentBytes
	msgBytes, err := json.Marshal(bindMsg)
	if err != nil {
		t.Fatalf("failed to marshal full message: %v", err)
	}

	ch.RouteMessage(client, msgBytes)

	// Read and check response
	select {
	case respBytes := <-client.Send:
		var resp map[string]interface{}
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if resp["op"] != "bindingStatus" {
			t.Errorf("expected op 'bindingStatus', got '%v'", resp["op"])
		}
		content, ok := resp["content"].(map[string]interface{})
		if !ok {
			t.Fatalf("response content is not a map: %v", resp["content"])
		}
		if content["state"] != "BOUND" {
			t.Errorf("expected state 'BOUND', got '%v'", content["state"])
		}
		if content["deviceId"] != "test-device" {
			t.Errorf("expected deviceId 'test-device', got '%v'", content["deviceId"])
		}
		if content["code"] != "BINDCODE" {
			t.Errorf("expected code 'BINDCODE', got '%v'", content["code"])
		}
		if _, ok := content["bindingId"]; !ok {
			t.Errorf("expected bindingId in response")
		}
		if _, ok := content["changeDate"]; !ok {
			t.Errorf("expected changeDate in response")
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive bind response in time")
	}
}

func TestHubOperations(t *testing.T) {
	hub := NewHub()
	if len(hub.channels) != 0 {
		t.Fatal("new hub should have no channels")
	}

	ch := hub.GetOrCreateChannel("new-channel")
	if ch == nil {
		t.Fatal("GetOrCreateChannel should return a channel")
	}
	if hub.channels["new-channel"] != ch {
		t.Fatal("channel not stored in hub correctly")
	}

	hub.DeleteChannel("new-channel")
	if _, ok := hub.channels["new-channel"]; ok {
		t.Fatal("channel not deleted from hub")
	}
}

func TestClientUnregister(t *testing.T) {
	hub := NewHub()
	ch := hub.GetOrCreateChannel("test-channel")

	client := &Client{Send: make(chan []byte, 1)}
	ch.Register(client)

	if len(ch.clients) != 1 {
		t.Fatal("client not registered")
	}

	ch.Unregister(client)

	if len(ch.clients) != 0 {
		t.Fatal("client not unregistered")
	}
	if _, ok := hub.channels["test-channel"]; ok {
		t.Fatal("channel should be deleted when last client unregisters")
	}
}

func TestRouteMessageEdgeCases(t *testing.T) {
	hub := NewHub()
	ch := hub.GetOrCreateChannel("edge-case-channel")

	mainClient := &Client{Send: make(chan []byte, 1), IsMain: true}
	ch.Register(mainClient)

	// Malformed JSON
	ch.RouteMessage(mainClient, []byte("not-a-json"))

	// Unhandled op
	ch.RouteMessage(mainClient, []byte("{\"op\": \"unknown\"}"))

	// Bind from non-main client (should be ignored)
	nonMainClient := &Client{Send: make(chan []byte, 1), IsMain: false}
	ch.Register(nonMainClient)
	bindMsg := []byte("{\"op\": \"bind\", \"content\": {}}")
	ch.RouteMessage(nonMainClient, bindMsg)

	select {
	case <-nonMainClient.Send:
		t.Fatal("non-main client should not receive a response for bind")
	case <-time.After(100 * time.Millisecond):
		// expected
	}
}

func TestMessageRoutingNoMainClient(t *testing.T) {
	hub := NewHub()
	ch := hub.GetOrCreateChannel("no-main-channel")

	client1 := &Client{Send: make(chan []byte, 1)}
	client2 := &Client{Send: make(chan []byte, 1)}
	ch.Register(client1)
	ch.Register(client2)

	// Message from a non-main client should not be routed anywhere
	msg := []byte("{\"op\": \"deviceStatus\"}")
	ch.RouteMessage(client1, msg)

	select {
	case <-client2.Send:
		t.Fatal("client2 should not receive message when no main client")
	case <-time.After(100 * time.Millisecond):
		// expected
	}
}
