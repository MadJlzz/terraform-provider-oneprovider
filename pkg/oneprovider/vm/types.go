package vm

import (
	"net/url"
	"strconv"
)

type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type InstanceCreateRequest struct {
	LocationId     int    `json:"location_id"`
	InstanceSizeId int    `json:"instance_size"`
	TemplateId     string `json:"template"`
	Hostname       string `json:"hostname"`
}

func (v *InstanceCreateRequest) UrlValues() url.Values {
	return url.Values{
		"location_id":   {strconv.Itoa(v.LocationId)},
		"instance_size": {strconv.Itoa(v.InstanceSizeId)},
		"template":      {v.TemplateId},
		"hostname":      {v.Hostname},
	}
}

type InstanceCreateResponse struct {
	Result   string `json:"result"`
	Response struct {
		Message   string `json:"message"`
		Id        string `json:"id"`
		IpAddress string `json:"ip_address"`
		Hostname  string `json:"hostname"`
		Password  string `json:"password"`
	} `json:"response"`
	Error *ApiError `json:"error"`
}

type ListTemplatesResponse struct {
	Templates []TemplateResponse `json:"response"`
}

type TemplateResponse struct {
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

type ListLocationsResponse struct {
	Result   string                        `json:"result"`
	Response map[string][]LocationResponse `json:"response"`
	Error    *ApiError                     `json:"error"`
}

type LocationResponse struct {
	Id             string               `json:"id"`
	Region         string               `json:"region"`
	Country        string               `json:"country"`
	City           string               `json:"city"`
	AvailableTypes []string             `json:"available_types"`
	AvailableSizes []int                `json:"available_sizes"`
	AvailableIPs   AvailableIPsResponse `json:"available_ips"`
}

type AvailableIPsResponse struct {
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
}

type InstanceDestroyRequest struct {
	VMId         string `json:"vm_id"`
	ConfirmClose bool   `json:"confirm_close"`
}

func (v *InstanceDestroyRequest) UrlValues() url.Values {
	return url.Values{
		"vm_id":         {v.VMId},
		"confirm_close": {strconv.FormatBool(v.ConfirmClose)},
	}
}

type InstanceDestroyResponse struct {
	Result   string `json:"result"`
	Response struct {
		Message                  string `json:"message"`
		UsageHours               string `json:"usageHours"`
		BandwidthOverusageInGB   string `json:"bandwidthOverusage"`
		BandwidthOverusageCost   string `json:"bandwidthOverusageCost"`
		AdditionalHoursForCharge string `json:"additionalHoursForCharge"`
	} `json:"response"`
	Error *ApiError `json:"error"`
}

type InstanceReadResponse struct {
	Result   string `json:"result"`
	Response struct {
		ServerInfo struct {
			IpAddress string `json:"ipaddress"`
			Hostname  string `json:"hostname"`
			City      string `json:"city"`
		} `json:"server_info"`
	} `json:"response"`
	Error *ApiError `json:"error"`
}

type InstanceUpdateRequest struct {
	VMId     string `json:"vm_id"`
	Hostname string `json:"hostname"`
}

func (v *InstanceUpdateRequest) HostnameUrlValues() url.Values {
	return url.Values{
		"vm_id":    {v.VMId},
		"hostname": {v.Hostname},
	}
}

type InstanceUpdateResponse struct {
	Result   string `json:"result"`
	Response struct {
		Message string `json:"message"`
	} `json:"response"`
	Error *ApiError `json:"error"`
}
