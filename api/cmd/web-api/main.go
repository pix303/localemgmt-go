package main

import (
	"log"
	"net/http"

	"fmt"
	"github.com/pix303/localemgmt-go/api/pkg/router"
)

func main() {
	r, err := router.NewRouter()
	if err != nil {
		log.Fatalf("Error on startup router: %v", err)
	}

	fmt.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", r.Router)
}
