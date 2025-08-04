package ssh

import "net/url"

type CreateSSHKeyRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

func (v *CreateSSHKeyRequest) UrlValues() url.Values {
	return url.Values{
		"key_name":  {v.Name},
		"key_value": {v.PublicKey},
	}
}

type CreateSSHKeyResponse struct {
	Response struct {
		Key struct {
			UUID  string `json:"uuid"`
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"key"`
	} `json:"response"`
}
