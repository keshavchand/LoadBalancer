package main

import (
	"net/http"
	"time"
)

func main() {
	lb := loadBalancer{}
	{ // The Load Balaner
		mux := http.NewServeMux()
		mux.HandleFunc("/", lb.HandleLoadBalaner)
		s := http.Server{}
		s.Addr = ":8000"
		s.Handler = mux
		s.ReadTimeout = 1 * time.Second
		s.IdleTimeout = 1 * time.Second
		s.WriteTimeout = 1 * time.Second
		go s.ListenAndServe()
	}
	{ // The Command And Control Server
		mux := http.NewServeMux()
		mux.HandleFunc("/", lb.HandleServerCNCRequest)
		s := http.Server{}
		s.Addr = ":7000"
		s.Handler = mux
		s.ReadTimeout = 1 * time.Second
		s.IdleTimeout = 1 * time.Second
		s.WriteTimeout = 1 * time.Second
		s.ListenAndServe()
	}
}
