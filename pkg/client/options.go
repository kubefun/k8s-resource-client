package client

type ClientOption func(*Client)

func WithResourceMode(mode ModeType) ClientOption {
	return func(c *Client) {
		c.ResourceMode = mode
	}
}

func WithNamespaceMode(mode ModeType) ClientOption {
	return func(c *Client) {
		c.NamespaceMode = mode
	}
}

func WithSkipSubjectAccessChecks(skip bool) ClientOption {
	return func(c *Client) {
		c.SkipSubjectAccessChecks = skip
	}
}
