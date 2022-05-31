package main

import (
	"fmt"
  "os"
	"log"
	"net/http"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("server <port>")
		return
	}
	_, err := strconv.Atoi(os.Args[1])
  if err != nil {
    log.Fatal(err)
    return
  }
  serverName := fmt.Sprintf("Server %s", os.Args[1])
	addr := ":" + os.Args[1]

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World from %s", serverName)
	})

	err = http.ListenAndServe(addr, nil)
	log.Fatal(err)
}
