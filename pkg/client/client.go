package client

type ModeType uint

const (
	Auto ModeType = iota
	Explicit
)

type Client struct {
	ResourceMode            ModeType
	NamespaceMode           ModeType
	SkipSubjectAccessChecks bool
}

func (c Client) Get() {
}

func NewClient(options ...ClientOption) *Client {
	c := &Client{
		ResourceMode:            Auto,
		NamespaceMode:           Auto,
		SkipSubjectAccessChecks: false,
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}
