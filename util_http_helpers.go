package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/context"
)

func GetIPFromRequest(r *http.Request) string {
	if fw := r.Header.Get("X-Forwarded-For"); fw != "" {
		// X-Forwarded-For has no port
		if i := strings.IndexByte(fw, ','); i >= 0 {
			return fw[:i]
		}
		return fw
	}

	// From net/http.Request.RemoteAddr:
	//   The HTTP server in this package sets RemoteAddr to an
	//   "IP:port" address before invoking a handler.
	// So we can ignore the case of the port missing.
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func CopyHttpRequest(r *http.Request) *http.Request {
	reqCopy := new(http.Request)
	*reqCopy = *r

	if r.Body != nil {
		defer r.Body.Close()

		// Buffer body data
		var bodyBuffer bytes.Buffer
		bodyBuffer2 := new(bytes.Buffer)

		io.Copy(&bodyBuffer, r.Body)
		*bodyBuffer2 = bodyBuffer

		// Create new ReadClosers so we can split output
		r.Body = ioutil.NopCloser(&bodyBuffer)
		reqCopy.Body = ioutil.NopCloser(bodyBuffer2)
	}

	return reqCopy
}

func CopyHttpResponse(r *http.Response) *http.Response {
	resCopy := new(http.Response)
	*resCopy = *r

	if r.Body != nil {
		defer r.Body.Close()

		// Buffer body data
		var bodyBuffer bytes.Buffer
		bodyBuffer2 := new(bytes.Buffer)

		io.Copy(&bodyBuffer, r.Body)
		*bodyBuffer2 = bodyBuffer

		// Create new ReadClosers so we can split output
		r.Body = ioutil.NopCloser(&bodyBuffer)
		resCopy.Body = ioutil.NopCloser(bodyBuffer2)
	}

	return resCopy
}

func RecordDetail(r *http.Request) bool {
	// Are we even checking?
	if !config.EnforceOrgDataDeailLogging {
		return config.AnalyticsConfig.EnableDetailedRecording
	}

	// We are, so get session data
	ses, found := context.GetOk(r, OrgSessionContext)
	if !found {
		// no session found, use global config
		return config.AnalyticsConfig.EnableDetailedRecording
	}

	// Session found
	return ses.(SessionState).EnableDetailedRecording
}
