package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
)

var urlRegexes = []*regexp.Regexp{
	genRegex(`^/v\d+\.\d+/containers/json\?limit=\d+$`),
	genRegex(`^/v\d+\.\d+/containers/[0-9a-fA-F]+/json$`),
	genRegex(`^/v\d+\.\d+/events\?filters=%7B%22type%22%3A%7B%22container%22%3Atrue%7D%7D$`),
	genRegex(`^/v\d+\.\d+/networks\?filters=`),
	genRegex(`^/v\d+\.\d+/services$`),
	genRegex(`^/v\d+\.\d+/tasks\?filters=`),
	genRegex(`^/v\d+\.\d+/version$`),
}

func genRegex(pattern string) (urlRegex *regexp.Regexp) {
	urlRegex, _ = regexp.Compile(pattern)
	return
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

type proxy struct {
	forwarder *http.Client
}

func (p *proxy) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.URL.Scheme != "" {
		msg := "unsupported protocol scheme: " + req.URL.String()
		http.Error(res, msg, http.StatusBadRequest)
		log.Println(msg)
		return
	}

	if req.Method != "GET" {
		msg := "unsupported method: " + req.Method
		http.Error(res, msg, http.StatusMethodNotAllowed)
		log.Println(msg)
		return
	}

	reqOk := false
	for _, re := range urlRegexes {
		if re.MatchString(req.URL.String()) {
			reqOk = true
			break
		}
	}

	if !reqOk {
		msg := "denied request to: " + req.URL.String()
		http.Error(res, msg, http.StatusForbidden)
		log.Println(msg)
		return
	}

	req.URL.Scheme = "http"
	req.URL.Host = "docker"

	// Request.RequestURI can't be set in client requests.
	// http://golang.org/src/pkg/net/http/client.go
	req.RequestURI = ""

	resp, err := p.forwarder.Do(req)
	if err != nil {
		http.Error(res, "Docker Socket Error", http.StatusBadGateway)
		log.Println("Docker Socket Error:", err)
		return
	}
	defer resp.Body.Close()

	res.WriteHeader(resp.StatusCode)
	copyHeaders(res.Header(), resp.Header)
	io.Copy(res, resp.Body)
}

func main() {
	handler := &proxy{
		forwarder: &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", "/var/run/docker.sock")
				},
			},
		},
	}

	if err := http.ListenAndServe(":2375", handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
