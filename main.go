package main

import (
	"io/ioutil"
	"net/http"
)

func cloneRequest(r *http.Request) *http.Request {
	method := r.Method
	body := r.Body

	path := r.URL.Path
	rawQuery := r.URL.RawQuery
	url := "https://github.com" + path + "?" + rawQuery

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	return req
}

func sendReqToUpstream(req *http.Request) []byte {
	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	return respBody
}

func handler(w http.ResponseWriter, r *http.Request) {
	req := cloneRequest(r)
	body := sendReqToUpstream(req)
	w.Write(body)
}

func main() {
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
