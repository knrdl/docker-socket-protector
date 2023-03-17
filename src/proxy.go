package main

import (
	"context"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
)

type FilterProxy struct {
	Rules       *[]ProfileRule
	Forwarder   *httputil.ReverseProxy
	LogRequests bool
}

func NewForwarder() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = "docker"
	}

	proxy := &httputil.ReverseProxy{
		Director: director,
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	return proxy
}

/*
a log message should not contain line breaks from arbitrary inputs as an attacker could otherwise fake log messages
*/
var newLineRegex = regexp.MustCompile(`\r?\n`)

func (p *FilterProxy) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	if p.LogRequests {
		msg := "request rule:  " + regexp.QuoteMeta(req.Method) + " " + regexp.QuoteMeta(req.URL.String())
		fmt.Fprintln(os.Stdout, newLineRegex.ReplaceAllString(msg, " "))
	}

	if req.URL.Scheme != "" {
		msg := "unsupported protocol scheme: " + req.URL.String()
		http.Error(res, html.EscapeString(msg), http.StatusBadRequest)
		fmt.Fprintln(os.Stderr, newLineRegex.ReplaceAllString(msg, " "))
		return
	}

	reqOk := false
	for _, rule := range *p.Rules {
		if rule.UrlRegex.MatchString(req.URL.String()) && rule.MethodRegex.MatchString(req.Method) {
			reqOk = true
			break
		}
	}

	if !reqOk {
		msg := "request denied: " + req.Method + " " + req.URL.String()
		http.Error(res, html.EscapeString(msg), http.StatusForbidden)
		fmt.Fprintln(os.Stderr, newLineRegex.ReplaceAllString(msg, " "))
		return
	}

	p.Forwarder.ServeHTTP(res, req)
}
