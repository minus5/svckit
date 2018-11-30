package api

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DigestRequest is a client for digest authentication requests
type DigestRequest struct {
	client             *http.Client
	username, password string
	nonceCount         nonceCount
}

type nonceCount int

func (nc nonceCount) String() string {
	c := int(nc)
	return fmt.Sprintf("%08x", c)
}

const authorization = "Authorization"
const nonce = "nonce"
const qop = "qop"
const realm = "realm"
const wwwAuthenticate = "Www-Authenticate"

var wanted = []string{nonce, qop, realm}

// NewDigestRequest makes a DigestRequest instance
func NewDigestRequest(username, password string) *DigestRequest {
	return &DigestRequest{
		client:   &http.Client{Timeout: 10 * time.Second},
		username: username,
		password: password,
	}
}

// Do does requests as http.Do does
func (r *DigestRequest) Do(req *http.Request) (*http.Response, error) {
	parts, err := r.makeParts(req)
	if err != nil {
		return nil, err
	}

	if parts != nil {
		req.Header.Set(authorization, r.makeAuthorization(req, parts))
	}

	return r.client.Do(req)
}

func (r *DigestRequest) makeParts(req *http.Request) (map[string]string, error) {
	authReq, err := http.NewRequest(req.Method, req.URL.String(), nil)
	resp, err := r.client.Do(authReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		return nil, nil
	}

	if len(resp.Header[wwwAuthenticate]) == 0 {
		return nil, fmt.Errorf("headers do not have %s", wwwAuthenticate)
	}

	headers := strings.Split(resp.Header[wwwAuthenticate][0], ",")
	parts := make(map[string]string, len(wanted))
	for _, r := range headers {
		for _, w := range wanted {
			if strings.Contains(r, w) {
				s := strings.Split(r, `"`)
				if len(s) > 1 {
					parts[w] = s[1]
				}
			}
		}
	}

	if len(parts) != len(wanted) {
		return nil, fmt.Errorf("header is invalid: %+v", parts)
	}

	return parts, nil
}

func getMD5(texts []string) string {
	h := md5.New()
	_, _ = io.WriteString(h, strings.Join(texts, ":"))
	return hex.EncodeToString(h.Sum(nil))
}

func (r *DigestRequest) getNonceCount() string {
	r.nonceCount++
	return r.nonceCount.String()
}

func (r *DigestRequest) makeAuthorization(req *http.Request, parts map[string]string) string {
	ha1 := getMD5([]string{r.username, parts[realm], r.password})
	ha2 := getMD5([]string{req.Method, req.URL.String()})
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	cnonce := fmt.Sprintf("%x", b)[:16]
	nc := r.getNonceCount()
	response := getMD5([]string{
		ha1,
		parts[nonce],
		nc,
		cnonce,
		parts[qop],
		ha2,
	})
	return fmt.Sprintf(
		`Digest username="%s", realm="%s", nonce="%s", uri="%s",  qop=%s, nc=%s, cnonce="%s", response="%s""`,
		r.username,
		parts[realm],
		parts[nonce],
		req.URL.String(),
		parts[qop],
		nc,
		cnonce,
		response,
	)
}
