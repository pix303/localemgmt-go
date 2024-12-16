package router

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func WelcomeHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome to Localemgmt")
}

func NewRouter() *httprouter.Router {
	router := httprouter.New()
	router.GET("/", WelcomeHandler)
	return router
}
