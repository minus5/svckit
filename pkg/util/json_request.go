// Doing retryable json request.
package util

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultRetries = 10
	MaxRetrySleep  = 10
)

type JsonRequest struct {
	url        string
	body       []byte
	Retries    int
	RetrySleep int64
	Header     http.Header
}

func NewJsonRequest(url string, body []byte) *JsonRequest {
	return &JsonRequest{
		url:        url,
		body:       body,
		Retries:    DefaultRetries,
		RetrySleep: MaxRetrySleep,
	}
}

func (r *JsonRequest) Do() (rsp []byte, err error, statusCode int) {
	var retryable bool
	for retry := 0; retry < r.Retries; retry++ {
		rsp, err, retryable, statusCode = r.do()
		if err == nil || !retryable {
			break
		} else {
			retryAfter := r.calcRetryInterval(retry)
			log.Printf("error: %s, will retry in %d seconds, retry no %d/%d", err, retryAfter, retry+1, r.Retries)
			time.Sleep(time.Duration(1e9 * retryAfter))
		}
	}
	return rsp, err, statusCode
}

func (r *JsonRequest) do() ([]byte, error, bool, int) {
	req, err := http.NewRequest("POST", r.url, bytes.NewReader(r.body))
	if err != nil {
		return nil, err, true, -1
	}
	if r.Header != nil {
		r.copyHeader(req.Header, r.Header)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.ContentLength = int64(len(r.body))
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err, true, -2
	}
	defer rsp.Body.Close()

	if rsp.StatusCode >= 300 && rsp.StatusCode < 500 {
		return nil, errors.New(fmt.Sprintf("%s response status code: %d", r.url, rsp.StatusCode)), false, rsp.StatusCode
	}
	if rsp.StatusCode < 200 || rsp.StatusCode >= 500 {
		return nil, errors.New(fmt.Sprintf("%s response status code: %d", r.url, rsp.StatusCode)), true, rsp.StatusCode
	}

	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err, true, -3
	}

	return rspBody, nil, true, rsp.StatusCode
}

func (r *JsonRequest) calcRetryInterval(retry int) int64 {
	retryAfter := int64(math.Exp(float64(retry)))
	if retryAfter > r.RetrySleep {
		retryAfter = r.RetrySleep
	}
	return retryAfter
}

func (r *JsonRequest) copyHeader(dst, src http.Header) {
	for k, vv := range src {
		if !strings.Contains(k, "Accept-Encoding") {
			for _, v := range vv {
				dst.Add(k, v)
			}
		}
	}
}
