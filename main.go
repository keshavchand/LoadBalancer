package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// TODO: Health check
func CreateProxy(site string) *httputil.ReverseProxy {
	url, err := url.Parse(site)
	if err != nil {
		log.Println(err)
		return nil
	}
	return httputil.NewSingleHostReverseProxy(url)
}

type loadBalancer struct {
	servers []*httputil.ReverseProxy
	last    int
}

func main() {
	lb := loadBalancer{
		servers: []*httputil.ReverseProxy{
			CreateProxy("http://localhost:8001"),
			CreateProxy("http://localhost:8002"),
			CreateProxy("http://localhost:8003"),
		},
	}

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
	lb.last++
	lb.last %= len(lb.servers)
	lb.servers[lb.last].ServeHTTP(w, r)
}
