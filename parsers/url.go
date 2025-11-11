package parsers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type (
	httpclient interface {
		Do(*http.Request) (*http.Response, error)
	}

	RemoteFileParser struct {
		u          *url.URL
		httpclient httpclient
		sep        string
	}

	RemoteFileParserOption func(*RemoteFileParser)
)

func WithHTTPClient(httpclient *http.Client) RemoteFileParserOption {
	return func(p *RemoteFileParser) {
		p.httpclient = httpclient
	}
}

func WithRemoteKeyValueSeparator(sep string) RemoteFileParserOption {
	return func(p *RemoteFileParser) {
		p.sep = sep
	}
}

func NewRemoteFileParser(u *url.URL, opts ...RemoteFileParserOption) *RemoteFileParser {
	p := &RemoteFileParser{
		u,
		http.DefaultClient,
		defaultKeyValueSeparator,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *RemoteFileParser) Parse() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req, err := http.NewRequest("get", p.u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpclient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("remote server returned unsuccessful status code: %d", resp.StatusCode)
	}

	kvp := NewKeyValueParser(resp.Body, WithKeyValueSeparator(p.sep))
	return kvp.Parse()
}
