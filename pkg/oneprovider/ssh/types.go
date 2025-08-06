package ssh

import "net/url"

type SshKeyCreateRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

func (v *SshKeyCreateRequest) UrlValues() url.Values {
	return url.Values{
		"key_name":  {v.Name},
		"key_value": {v.PublicKey},
	}
}

type SshKeyCreateResponse struct {
	Response struct {
		Key struct {
			Uuid  string `json:"uuid"`
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"key"`
	} `json:"response"`
}

type SshKeyReadResponse struct {
	Uuid  string `json:"uuid"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SshKeyListResponse struct {
	Response struct {
		SshKeys []SshKeyReadResponse `json:"keys"`
	} `json:"response"`
}

type SshKeyUpdateRequest struct {
	Uuid      string `json:"uuid"`
	Name      string `json:"name"`
	PublicKey string `json:"value"`
}

func (v *SshKeyUpdateRequest) UrlValues() url.Values {
	return url.Values{
		"ssh_key":   {v.Uuid},
		"key_name":  {v.Name},
		"key_value": {v.PublicKey},
	}
}

type SshKeyUpdateResponse struct {
	Response struct {
		SshKeys []struct {
			Name      string `json:"name"`
			PublicKey string `json:"value"`
		}
	}
}
