package oneprovider

type API interface {
	ListTemplates() []string
}

type service struct {
	host      string
	apiKey    string
	clientKey string
}

func NewService(host, apiKey, clientKey string) (API, error) {
	return &service{
		host:      host,
		apiKey:    apiKey,
		clientKey: clientKey,
	}, nil
}

func (s *service) ListTemplates() []string {
	return nil
}
