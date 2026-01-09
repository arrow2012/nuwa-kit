package social

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/arrow2012/nuwa-kit/pkg/json"
	"github.com/arrow2012/nuwa-kit/pkg/metric"
	"github.com/sony/gobreaker"
)

type WeComProvider struct {
	CorpID  string
	AgentID string
	Secret  string
	cb      *gobreaker.CircuitBreaker
	mu      sync.RWMutex
}

type WeComUserInfo struct {
	UserID string `json:"UserId"`
	OpenID string `json:"OpenId"` // Used if not part of corp
	Name   string `json:"name"`   // Often empty in initial auth, needs separate User Get
	Email  string `json:"email"`
}

func NewWeComProvider(corpID, agentID, secret string) *WeComProvider {
	// Configure Circuit Breaker
	settings := gobreaker.Settings{
		Name:        "WeCom",
		MaxRequests: 5,                // Max concurrent requests in half-open state
		Interval:    60 * time.Second, // Cyclic period of closed state (to clear counts)
		Timeout:     30 * time.Second, // Open state duration
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip if failure ratio > 40% and at least 5 requests
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.4
		},
	}

	return &WeComProvider{
		CorpID:  corpID,
		AgentID: agentID,
		Secret:  secret,
		cb:      gobreaker.NewCircuitBreaker(settings),
	}
}

// UpdateCircuitBreaker updates the circuit breaker settings
func (p *WeComProvider) UpdateCircuitBreaker(maxRequests uint32, interval, timeout float64, ratio float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	settings := gobreaker.Settings{
		Name:        "WeCom",
		MaxRequests: maxRequests,
		Interval:    time.Duration(interval * float64(time.Second)),
		Timeout:     time.Duration(timeout * float64(time.Second)),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= ratio
		},
	}
	p.cb = gobreaker.NewCircuitBreaker(settings)
}

// GenerateLoginURL constructs the QR Connect URL
// https://open.work.weixin.qq.com/wwopen/sso/qrConnect?appid=CORPID&agentid=AGENTID&redirect_uri=REDIRECT_URI&state=STATE
func (p *WeComProvider) GenerateLoginURL(redirectURI, state string) string {
	baseURL := "https://open.work.weixin.qq.com/wwopen/sso/qrConnect"
	u, _ := url.Parse(baseURL)
	q := u.Query()
	q.Set("appid", p.CorpID)
	q.Set("agentid", p.AgentID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String()
}

// GetUserInfo processes the callback code code and retrieves UserID
func (p *WeComProvider) GetUserInfo(code string) (userInfo *WeComUserInfo, err error) {
	start := time.Now()
	defer func() {
		status := "success"
		if err != nil {
			status = "failed"
		}
		// Assuming generic ExternalAPIDuration exists in kit metric, if not we need check.
		// Previous context says I added it.
		metric.ExternalAPIDuration.WithLabelValues("wecom", "get_user_info", status).Observe(time.Since(start).Seconds())
	}()

	// Check configuration
	if p.CorpID == "" || p.Secret == "" {
		return nil, fmt.Errorf("wecom provider not configured")
	}

	// Use Circuit Breaker
	// The return value of Execute is (interface{}, error)
	// Need to lock read access to cb because it might be replaced
	p.mu.RLock()
	cb := p.cb
	p.mu.RUnlock()

	res, err := cb.Execute(func() (interface{}, error) {
		// 1. Get Access Token
		tokenURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", p.CorpID, p.Secret)
		resp, err := http.Get(tokenURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get access token: %v", err)
		}
		defer resp.Body.Close()

		var tokenResp struct {
			ErrCode     int    `json:"errcode"`
			AccessToken string `json:"access_token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			return nil, err
		}
		if tokenResp.ErrCode != 0 {
			// If error is from WeCom API configuration , maybe we shouldn't trip?
			// But for now treat as failure.
			return nil, fmt.Errorf("wecom token error: %d", tokenResp.ErrCode)
		}

		// 2. Get User ID from Code
		userInfoURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/getuserinfo?access_token=%s&code=%s", tokenResp.AccessToken, code)
		resp, err = http.Get(userInfoURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get user info: %v", err)
		}
		defer resp.Body.Close()

		var userResp struct {
			ErrCode int    `json:"errcode"`
			UserID  string `json:"UserId"`
			OpenID  string `json:"OpenId"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
			return nil, err
		}
		if userResp.ErrCode != 0 {
			return nil, fmt.Errorf("wecom user info error: %d", userResp.ErrCode)
		}

		return &WeComUserInfo{
			UserID: userResp.UserID,
			OpenID: userResp.OpenID,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return res.(*WeComUserInfo), nil
}
