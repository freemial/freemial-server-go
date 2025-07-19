package api

import (
	"encoding/json"
	"net/http"

	"github.com/freemial/freemial-server-go/internal/websocket"
)

type Binding struct {
	BindingId  int    `json:"bindingId"`
	DeviceId   string `json:"deviceId"`
	UserId     int    `json:"userId"`
	State      string `json:"state"`
	ChangeDate string `json:"changeDate"`
	Code       string `json:"code"`
	Name       string `json:"name"`
}

func GetDeviceBindings(hub *websocket.Hub, w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE,GET,HEAD,OPTIONS,PUT,POST,PATCH")
	if r.Method == "OPTIONS" {
		return
	}

	channelNames := hub.GetChannelNames()

	bindings := make([]Binding, len(channelNames))
	for i, name := range channelNames {
		bindings[i] = Binding{
			BindingId:  i,
			DeviceId:   name,
			UserId:     1,
			State:      "BOUND",
			ChangeDate: "changeDate",
			Code:       "code",
			Name:       name,
		}
	}

	ret, err := json.Marshal(bindings)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(ret)
}
