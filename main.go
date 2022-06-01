package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// TODO: Health check
func CreateProxy(site string) (ReverseProxy, error) {
	var rp ReverseProxy
	url, err := url.Parse(site)
	if err != nil {
		return rp, err
	}
	rp.proxy = httputil.NewSingleHostReverseProxy(url)
	rp.url = url
	rp.valid = true
	return rp, nil
}

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

func main() {
	backends := []string{
		"http://localhost:8001",
		"http://localhost:8002",
		"http://localhost:8003",
		"http://localhost:8004",
		"http://localhost:8005",
	}

	lb := loadBalancer{}
	for _, s := range backends {
		rp, err := CreateProxy(s)
		if err != nil {
			log.Println(err)
			continue
		}
		lb.servers = append(lb.servers, rp)
	}
	lb.SetHealthCheck()

	mux := http.NewServeMux()
	mux.HandleFunc("/", lb.HandleLoadBalaner)

	s := http.Server{}
	s.Addr = ":8000"
	s.Handler = mux
	s.ReadTimeout = 1 * time.Second
	s.IdleTimeout = 1 * time.Second
	s.WriteTimeout = 1 * time.Second

	s.ListenAndServe()
}

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

func (lb *loadBalancer) SetHealthCheck() {
	for i := 0; i < len(lb.servers); i++ {
		lb.servers[i].proxy.ErrorHandler = lb.ErrorHandler
	}
}
