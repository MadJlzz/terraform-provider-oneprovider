package vm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

type NumericString string

func (v NumericString) String() string {
	return string(v)
}

func (v *NumericString) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("invalid numeric string JSON: %w", err)
	}

	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("invalid numeric string JSON: trailing data")
		}
		return fmt.Errorf("invalid numeric string JSON: trailing data: %w", err)
	}

	switch value := value.(type) {
	case string:
		if !isNonNegativeDecimalInteger(value) {
			return fmt.Errorf("invalid numeric string %q", value)
		}
		*v = NumericString(value)
		return nil
	case json.Number:
		valueString := value.String()
		if !isNonNegativeDecimalInteger(valueString) {
			return fmt.Errorf("invalid numeric string number %q", valueString)
		}
		*v = NumericString(valueString)
		return nil
	default:
		return fmt.Errorf("invalid numeric string JSON type %T", value)
	}
}

type APIID string

func (v APIID) String() string {
	return string(v)
}

func (v *APIID) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("invalid API ID JSON: %w", err)
	}

	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("invalid API ID JSON: trailing data")
		}
		return fmt.Errorf("invalid API ID JSON: trailing data: %w", err)
	}

	switch value := value.(type) {
	case string:
		if value == "" {
			return fmt.Errorf("invalid API ID: empty string")
		}
		*v = APIID(value)
		return nil
	case json.Number:
		valueString := value.String()
		if !isNonNegativeDecimalInteger(valueString) {
			return fmt.Errorf("invalid API ID number %q", valueString)
		}
		*v = APIID(valueString)
		return nil
	default:
		return fmt.Errorf("invalid API ID JSON type %T", value)
	}
}

func isNonNegativeDecimalInteger(value string) bool {
	if value == "" {
		return false
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

type TemplatesListResponse struct {
	Templates []TemplateReadResponse `json:"response"`
}

type TemplateReadResponse struct {
	Id      APIID  `json:"id"`
	Name    string `json:"name"`
	Size    string `json:"size"`
	Display struct {
		Name        string `json:"name"`
		Display     string `json:"display"`
		Description string `json:"description"`
		Oca         int    `json:"oca"`
	} `json:"display"`
}

type LocationsListResponse struct {
	Response map[string][]LocationReadResponse `json:"response"`
}

type LocationReadResponse struct {
	Id             APIID    `json:"id"`
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
			Plan      string `json:"plan"`
			Template  string `json:"template"`
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
		urlValues.Add(fmt.Sprintf("ssh_keys[%d]", idx), key)
	}
	return urlValues
}

type InstanceCreateResponse struct {
	Response struct {
		Message   string `json:"message"`
		Id        APIID  `json:"id"`
		IpAddress string `json:"ip_address"`
		Hostname  string `json:"hostname"`
		Password  string `json:"password"`
	} `json:"response"`
}

type InstanceHostnameUpdateRequest struct {
	VmId     string `json:"vm_id"`
	Hostname string `json:"hostname"`
}

func (v *InstanceHostnameUpdateRequest) UrlValues() url.Values {
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

type SizesListResponse struct {
	Response []SizeReadResponse `json:"response"`
}

type SizeReadResponse struct {
	Id    APIID         `json:"id"`
	Name  string        `json:"name"`
	Type  string        `json:"type"`
	Cores NumericString `json:"cores"`
	RAM   NumericString `json:"ram"`
	Disk  NumericString `json:"hdd"`
}
