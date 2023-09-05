package oneprovider

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

type Location int64

const (
	NorthAmerica Location = iota
	Europe
	Asia
	Oceania
	SouthAmerica
	Africa
)

func (l Location) String() string {
	switch l {
	case NorthAmerica:
		return "North America"
	case Europe:
		return "Europe"
	case Asia:
		return "Asia"
	case Oceania:
		return "Oceania"
	case SouthAmerica:
		return "South America"
	case Africa:
		return "Africa"
	}
	return "unknown"
}

type ListLocationsResponse struct {
	Locations map[string][]LocationResponse `json:"response"`
}

type LocationResponse struct {
	Id             string   `json:"id"`
	Region         string   `json:"region"`
	Country        string   `json:"country"`
	City           string   `json:"city"`
	AvailableTypes []string `json:"available_types"`
	AvailableSizes []int    `json:"available_sizes"`
}
