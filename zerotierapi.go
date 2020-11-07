package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const zeroTierApiUrl = "https://my.zerotier.com/api"
const addMemberPayload = "{\"hidden\":false,\"config\":{\"authorized\":true}}"

type ZeroTierApi struct {
	accessToken string
}

func NewZTApi(token string) *ZeroTierApi {
	return &ZeroTierApi{
		accessToken: token,
	}
}

func (api *ZeroTierApi) AddMember(network string, user string) (bool, error) {
	req, err := http.NewRequest("POST",
		zeroTierApiUrl+fmt.Sprintf("/network/%s/member/%s", network, user),
		strings.NewReader(addMemberPayload))
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "bearer "+api.accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("ZT Api: Failed to add user %s to network %s (status %s)\n", user, network, resp.Status)
	}

	return resp.StatusCode == http.StatusOK, nil
}
