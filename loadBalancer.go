package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
)

type loadBalancer struct {
	sync.RWMutex
	servers []ReverseProxy
	last    int
}

type ReverseProxy struct {
	proxy *httputil.ReverseProxy
	url   *url.URL
	valid bool
}

const ServerCNCRequestName string = "RequestType"
const ServerCNCRequestAdd string = "add"
const ServerCNCRequestRemove string = "remove"
const ServerCNCRequestAddr string = "Addr"

func (lb *loadBalancer) HandleLoadBalaner(w http.ResponseWriter, r *http.Request) {
	server := lb.GetValidServer()
	if server == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	server.ServeHTTP(w, r)
}

func (lb *loadBalancer) GetValidServer() *httputil.ReverseProxy {
	lb.RLock()
	defer lb.RUnlock()
	for i := 0; i < len(lb.servers); i++ {
		lb.last++
		lb.last %= len(lb.servers)
		if lb.servers[lb.last].valid {
			log.Println("returning", lb.servers[lb.last].url.Host)
			return lb.servers[lb.last].proxy
		}
	}
	log.Println("no valid server to return")
	return nil
}

func (lb *loadBalancer) ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	lb.Lock()
	defer lb.Unlock()
	log.Println(r.URL.Host, ":", err)
	for i := 0; i < len(lb.servers); i++ {
		if lb.servers[i].url.Host == r.URL.Host {
			lb.servers[i].valid = false
		}
	}
}

func (lb *loadBalancer) HandleServerCNCRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse Form
	// Get the type of request begin sent
	r.ParseForm()
	reqType := r.Form[ServerCNCRequestName]
	if len(reqType) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	addr := r.Form[ServerCNCRequestAddr]
	if len(addr) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Note: We only care about the first one
	req := reqType[0]
	switch strings.ToLower(req) {
	case ServerCNCRequestAdd:
		lb.AddServer(addr[0])
	case ServerCNCRequestRemove:
		lb.RemoveServer(addr[0])
	}
}

func (lb *loadBalancer) AddServer(addr string) {
	var rp ReverseProxy
	url, err := url.Parse(addr)
	if err != nil { }
	rp.proxy = httputil.NewSingleHostReverseProxy(url)
	rp.url = url
	rp.valid = true

	lb.Lock()
	defer lb.Unlock()
	// TODO: Find and insert in the list
	for i := 0; i < len(lb.servers); i++ {
		if lb.servers[i].url.String() == addr {
			lb.servers[i] = rp
			return
		}
	}

	lb.servers = append(lb.servers, rp)
}

func (lb *loadBalancer) RemoveServer(addr string) {
	var rp ReverseProxy
	url, err := url.Parse(addr)
	if err != nil { }
	rp.proxy = httputil.NewSingleHostReverseProxy(url)
	rp.url = url
	rp.valid = true

	lb.Lock()
	defer lb.Unlock()
	// TODO: Find and insert in the list
	for i := 0; i < len(lb.servers); i++ {
		if lb.servers[i].url.String() == addr {
			lb.servers[i].valid = false
			return
		}
	}
}
