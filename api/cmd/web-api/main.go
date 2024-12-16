package main

import (
	"github.com/pix303/localemgmt-go/api/pkg/router"
	"net/http"
)

func main() {
	r := router.NewRouter()
	http.ListenAndServe(":8080", r)
}
