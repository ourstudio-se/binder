package parsers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type httpclient interface {
	Do(*http.Request) (*http.Response, error)
}

type RemoteFileParser struct {
	u          *url.URL
	httpclient httpclient
}

func NewRemoteFileParser(u *url.URL) *RemoteFileParser {
	return &RemoteFileParser{
		u,
		http.DefaultClient,
	}
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

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("remote server returned unsuccessful status code: %d", resp.StatusCode)
	}

	kvp := NewKeyValueParser(resp.Body)
	return kvp.Parse()
}
