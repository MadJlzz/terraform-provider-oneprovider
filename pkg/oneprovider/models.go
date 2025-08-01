package oneprovider

type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ListVMTemplatesResponse struct {
	Templates []VMTemplateResponse `json:"response"`
}

type VMTemplateResponse struct {
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

type ListVMLocationsResponse struct {
	Result   string                          `json:"result"`
	Response map[string][]VMLocationResponse `json:"response"`
	Error    *ApiError                       `json:"error"`
}

type VMLocationResponse struct {
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
