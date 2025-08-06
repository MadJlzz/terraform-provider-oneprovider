package vm

import (
	"fmt"
	"net/url"
	"strconv"
)

type TemplatesListResponse struct {
	Templates []TemplateReadResponse `json:"response"`
}

type TemplateReadResponse struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Size    string `json:"size"`
	Display struct {
		Name        string `json:"name"`
		Display     string `json:"display"`
		Description string `json:"description"`
		Oca         int    `json:"oca"`
	}
}

type LocationsListResponse struct {
	Response map[string][]LocationReadResponse `json:"response"`
}

type LocationReadResponse struct {
	Id             string   `json:"id"`
	Region         string   `json:"region"`
	Country        string   `json:"country"`
	City           string   `json:"city"`
	AvailableTypes []string `json:"available_types"`
	AvailableSizes []int    `json:"available_sizes"`
	AvailableIPs   struct {
		IPv4 string `json:"ipv4"`
		IPv6 string `json:"ipv6"`
	} `json:"available_ips"`
}

type InstanceReadResponse struct {
	Response struct {
		ServerInstall bool `json:"server_install"`
		ServerInfo    struct {
			IpAddress string `json:"ipaddress"`
			Hostname  string `json:"hostname"`
			City      string `json:"city"`
		} `json:"server_info"`
		ServerState struct {
			Status string `json:"status"`
			State  string `json:"state"`
		} `json:"server_state"`
	} `json:"response"`
}

type InstanceCreateRequest struct {
	LocationId     int      `json:"location_id"`
	InstanceSizeId int      `json:"instance_size"`
	TemplateId     string   `json:"template"`
	Hostname       string   `json:"hostname"`
	SshKeys        []string `json:"ssh_keys"`
}

func (v *InstanceCreateRequest) UrlValues() url.Values {
	urlValues := url.Values{
		"location_id":   {strconv.Itoa(v.LocationId)},
		"instance_size": {strconv.Itoa(v.InstanceSizeId)},
		"template":      {v.TemplateId},
		"hostname":      {v.Hostname},
	}
	for idx, key := range v.SshKeys {
		//urlValues.Add(fmt.Sprintf("keys[%d]", idx), key)
		urlValues.Add(fmt.Sprintf("ssh_keys[%d]", idx), key)
		//urlValues.Add("ssh_keys", fmt.Sprintf("[%s]", key))
	}
	return urlValues
}

type InstanceCreateResponse struct {
	Response struct {
		Message   string `json:"message"`
		Id        string `json:"id"`
		IpAddress string `json:"ip_address"`
		Hostname  string `json:"hostname"`
		Password  string `json:"password"`
	} `json:"response"`
}

type InstanceHostnameUpdateRequest struct {
	VmId     string `json:"vm_id"`
	Hostname string `json:"hostname"`
}

func (v *InstanceHostnameUpdateRequest) HostnameUrlValues() url.Values {
	return url.Values{
		"vm_id":    {v.VmId},
		"hostname": {v.Hostname},
	}
}

type InstanceDestroyRequest struct {
	VmId         string `json:"vm_id"`
	ConfirmClose bool   `json:"confirm_close"`
}

func (v *InstanceDestroyRequest) UrlValues() url.Values {
	return url.Values{
		"vm_id":         {v.VmId},
		"confirm_close": {strconv.FormatBool(v.ConfirmClose)},
	}
}
