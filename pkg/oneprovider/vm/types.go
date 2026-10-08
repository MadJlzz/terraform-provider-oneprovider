package vm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// StringOrNumber preserves response values which OneProvider returns either as
// JSON strings or JSON numbers. JSON null is represented by the empty string.
// Business validation belongs at the use site, not at this wire seam.
type StringOrNumber string

func (v StringOrNumber) String() string {
	return string(v)
}

func (v *StringOrNumber) UnmarshalJSON(data []byte) error {
	if isJSONNull(data) {
		return nil
	}
	value, err := decodeStringOrNumber(data)
	if err != nil {
		return fmt.Errorf("invalid string/number JSON: %w", err)
	}
	*v = StringOrNumber(value)
	return nil
}

// NumericString is kept as a source-compatible name for callers of earlier
// provider versions. It has the same tolerant wire semantics as StringOrNumber.
type NumericString = StringOrNumber

var jsonNull = []byte("null")

func isJSONNull(data []byte) bool {
	return bytes.Equal(data, jsonNull)
}

// decodeStringOrNumber inspects the raw JSON value instead of decoding it into
// an interface, so numbers keep their exact wire text. Callers handle null.
func decodeStringOrNumber(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty JSON value")
	}

	switch c := data[0]; {
	case c == '"':
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return "", err
		}
		return value, nil
	case c == '-' || ('0' <= c && c <= '9'):
		// UnmarshalJSON may be called directly, so the number literal is
		// validated here rather than relying on json.Unmarshal having done it.
		if !json.Valid(data) {
			return "", fmt.Errorf("invalid JSON number")
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported JSON value starting with %q", c)
	}
}

// APIID is opaque response text. OneProvider may encode it as a JSON string,
// number, or null; whether an operation requires a numeric ID is validated by
// that operation instead of this response decoder.
type APIID string

func (v APIID) String() string {
	return string(v)
}

func (v *APIID) UnmarshalJSON(data []byte) error {
	if isJSONNull(data) {
		return nil
	}
	value, err := decodeStringOrNumber(data)
	if err != nil {
		return fmt.Errorf("invalid API ID JSON: %w", err)
	}
	*v = APIID(value)
	return nil
}

// BoolOrString accepts the two representations used by /vm/info for
// server_install: a JSON boolean and the documented JSON string "false".
// JSON null leaves the value unchanged, matching a plain bool field.
type BoolOrString bool

func (v *BoolOrString) UnmarshalJSON(data []byte) error {
	if isJSONNull(data) {
		return nil
	}

	value := string(data)
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("invalid bool/string JSON: %w", err)
		}
	}

	switch value {
	case "true":
		*v = true
	case "false":
		*v = false
	default:
		return fmt.Errorf("invalid bool/string JSON value")
	}
	return nil
}

type TemplatesListResponse struct {
	Templates []TemplateReadResponse `json:"response"`
}

type TemplateDisplayReadResponse struct {
	Name        string         `json:"name"`
	Display     string         `json:"display"`
	Description string         `json:"description"`
	Oca         StringOrNumber `json:"oca"`
}

type TemplateReadResponse struct {
	Id      APIID                       `json:"id"`
	Name    string                      `json:"name"`
	Size    StringOrNumber              `json:"size"`
	Display TemplateDisplayReadResponse `json:"display"`
}

type LocationsListResponse struct {
	Response map[string][]LocationReadResponse `json:"response"`
}

type LocationReadResponse struct {
	Id             APIID            `json:"id"`
	Region         string           `json:"region"`
	Country        string           `json:"country"`
	City           string           `json:"city"`
	AvailableTypes []string         `json:"available_types"`
	AvailableSizes []StringOrNumber `json:"available_sizes"`
	AvailableIPs   struct {
		IPv4 string `json:"ipv4"`
		IPv6 string `json:"ipv6"`
	} `json:"available_ips"`
}

type InstanceReadResponse struct {
	Response struct {
		ServerInstall BoolOrString `json:"server_install"`
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
	Id    APIID          `json:"id"`
	Name  string         `json:"name"`
	Type  string         `json:"type"`
	Cores StringOrNumber `json:"cores"`
	RAM   StringOrNumber `json:"ram"`
	Disk  StringOrNumber `json:"hdd"`
}
