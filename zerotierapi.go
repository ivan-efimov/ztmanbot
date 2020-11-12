package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const zeroTierApiUrl = "https://my.zerotier.com/api"

var (
	InvalidNodeId    = errors.New("invalid NodeID format")
	InvalidNetworkId = errors.New("invalid NetworkID format")
)

// This struct is not complete yet
type MemberConfigEditablePart struct {
	Authorized    bool     `json:"authorized"`
	IpAssignments []string `json:"ipAssignments"`
}

// This struct is not complete yet
type MemberEditablePart struct {
	Hidden      bool                     `json:"hidden"`
	Name        string                   `json:"name,omitempty"`
	Description string                   `json:"description,omitempty"`
	Config      MemberConfigEditablePart `json:"config,omitempty"`
}

// This struct is not complete yet
type MemberInfo struct {
	MemberEditablePart
	NodeID          string `json:"nodeId"`
	Online          bool   `json:"online"`
	PhysicalAddress string `json:"physicalAddress"`
	ClientVersion   string `json:"clientVersion"`
}

type ZeroTierApi struct {
	accessToken    string
	defaultNetwork string

	nodeIdRegEx    *regexp.Regexp
	networkIdRegEx *regexp.Regexp
}

func NewZTApi(token string, defaultNetwork string) *ZeroTierApi {
	return &ZeroTierApi{
		accessToken:    token,
		defaultNetwork: defaultNetwork,
		nodeIdRegEx:    regexp.MustCompile("^[0-9a-f]{10}$"),
		networkIdRegEx: regexp.MustCompile("^[0-9a-f]{16}$"),
	}
}

func (api *ZeroTierApi) AuthMember(networkId string, nodeId string, shortName string, description string) (bool, error) {
	if !api.networkIdRegEx.MatchString(networkId) {
		return false, InvalidNetworkId
	}
	if !api.nodeIdRegEx.MatchString(nodeId) {
		return false, InvalidNodeId
	}
	memberChanges := &MemberEditablePart{
		Hidden:      false,
		Name:        shortName,
		Description: description,
		Config: MemberConfigEditablePart{
			Authorized: true,
		},
	}
	jsonBytes, err := json.Marshal(memberChanges)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST",
		zeroTierApiUrl+fmt.Sprintf("/network/%s/member/%s", networkId, nodeId),
		strings.NewReader(string(jsonBytes)))
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
		log.Printf("ZeroTierApi.AuthMember: Failed to auth user %s in network %s (status %s)\n", nodeId, networkId, resp.Status)
	}

	return resp.StatusCode == http.StatusOK, nil
}

func (api *ZeroTierApi) UnauthMemberByID(networkId string, nodeId string) (bool, error) {
	if !api.networkIdRegEx.MatchString(networkId) {
		return false, InvalidNetworkId
	}
	if !api.nodeIdRegEx.MatchString(nodeId) {
		return false, InvalidNodeId
	}
	memberChanges := &MemberEditablePart{
		Hidden: false,
		Config: MemberConfigEditablePart{
			Authorized: false,
		},
	}
	jsonBytes, err := json.Marshal(memberChanges)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST",
		zeroTierApiUrl+fmt.Sprintf("/network/%s/member/%s", networkId, nodeId),
		strings.NewReader(string(jsonBytes)))
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
		log.Printf("ZeroTierApi.UnauthMember: Failed to unauth user %s in network %s (status %s)\n", nodeId, networkId, resp.Status)
	}

	return resp.StatusCode == http.StatusOK, nil
}

func (api *ZeroTierApi) ListMembers(networkId string) ([]*MemberInfo, error) {
	if !api.networkIdRegEx.MatchString(networkId) {
		return nil, InvalidNetworkId
	}
	req, err := http.NewRequest("GET",
		zeroTierApiUrl+fmt.Sprintf("/network/%s/member", networkId),
		nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "bearer "+api.accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	members := make([]*MemberInfo, 0)

	if resp.StatusCode != http.StatusOK {
		log.Printf("ZeroTierApi.ListMembers: Failed to get list of members of network %s (status %s)\n", networkId, resp.Status)
		return nil, nil
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&members)
	if err != nil {
		return nil, err
	}

	return members, nil
}
