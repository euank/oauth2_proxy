package providers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"
)

type WobProvider struct {
	*OIDCProvider

	url    *url.URL
	groups []string
}

func NewWobProvider(p *ProviderData) *WobProvider {
	op := NewOIDCProvider(p)
	op.ProviderName = "Wobscale Accounts"
	if op.Scope == "" {
		op.Scope = "user:email"
	}
	return &WobProvider{op, nil, []string{}}
}

func (p *WobProvider) SetRequiredGroups(groups []string) {
	p.Scope += " group:check_membership"
	p.groups = groups
}

func (p *WobProvider) SetWobURL(u *url.URL) {
	p.url = u
}

func (p *WobProvider) ValidateSession(s *SessionState) bool {
	isIn, err := p.inGroups(s.AccessToken)
	if err != nil {
		log.Printf("[wob] error checking group membership: %v", err)
	}
	return isIn
}

func (p *WobProvider) inGroups(accessToken string) (bool, error) {
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	if p.url == nil {
		// need a validate url to check
		return false, fmt.Errorf("no validate url specified")
	}
	for _, g := range p.groups {
		ep := *p.url
		ep.Path = path.Join(ep.Path, fmt.Sprintf("group/%s/membership", g))
		req, _ := http.NewRequest("GET", ep.String(), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		req.Header.Set("Accept", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Printf("[wob] error requesting group membership: %v", err)
			return false, err
		}
		var m membershipResponse
		if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
			return false, fmt.Errorf("wob: unable to decode membership response: %v", err)
		}
		log.Printf("decided into %v", m)
		if !m.Member {
			return false, nil
		}
	}
	return true, nil
}

type membershipResponse struct {
	Member bool `json:"member"`
}
