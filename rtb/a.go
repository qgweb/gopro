package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Query().Get("prod"))
		fmt.Println(r.URL.Query().Get("showType"))
		fmt.Println(r.URL.Query().Get("mid"))
	})
	http.ListenAndServe(":8081", nil)
}
