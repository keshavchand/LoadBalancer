package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

func Server(i int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    forward := r.Header.Get("X-Forwarded-For")
    if forward != "" {
      fmt.Println(forward)
    }
		fmt.Fprintf(os.Stdout, "Hello World from %d", i)
		fmt.Fprintf(w, "Hello World from %d", i)
	})
	s := http.Server{}
	s.Addr = fmt.Sprintf(":%d", i)
	s.Handler = mux
	err := s.ListenAndServe()
	log.Fatal(err)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("server <count>")
	}

	base := 8001
	count, err := strconv.Atoi(os.Args[1])
  if err != nil {
    log.Fatal(err)
  }
	var wg sync.WaitGroup
	defer wg.Wait()
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			log.Println("Servering at", i)
			defer wg.Done()
			Server(i)
		}(base + i)
	}
}
