package api

import (
	"encoding/json"
	"net/http"
)

type Credentials struct {
	UserId          int    `json:"userId"`
	AccessToken     string `json:"accessToken"`
	RefreshToken    string `json:"refreshToken"`
	ExpiresIn       int    `json:"expiresIn"`
	FirstLogin      bool   `json:"firstLogin"`
	AccountProvider string `json:"accountProvider"`
	ChangeDate      string `json:"changeDate"`
	ErrorCode       int    `json:"errorCode"`
}

func Login(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE,GET,HEAD,OPTIONS,PUT,POST,PATCH")
	if r.Method == "OPTIONS" {
		return
	}
	ret, err := json.Marshal(Credentials{
		UserId:          1,
		AccessToken:     "accessToken",
		RefreshToken:    "refreshToken",
		ExpiresIn:       3600,
		FirstLogin:      true,
		AccountProvider: "accountProvider",
		ChangeDate:      "changeDate",
		ErrorCode:       0,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(ret)
}
