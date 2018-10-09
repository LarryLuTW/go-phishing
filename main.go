package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func replaceURLInResp(body []byte, header http.Header) []byte {
	// detect is html or not
	contentType := header.Get("Content-Type")
	isHTML := strings.Contains(contentType, "text/html")

	// if NOT, return same body
	if !isHTML {
		return body
	}

	bodyStr := string(body)
	bodyStr = strings.Replace(bodyStr, "https://github.com", "http://localhost:8080", -1)

	re, err := regexp.Compile(`http://localhost:8080(.*)\.git`)
	if err != nil {
		panic(err)
	}
	bodyStr = re.ReplaceAllString(bodyStr, `https://github.com$1.git`)

	return []byte(bodyStr)
}

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

func sendReqToUpstream(req *http.Request) ([]byte, http.Header) {
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

	return respBody, resp.Header
}

func handler(w http.ResponseWriter, r *http.Request) {
	req := cloneRequest(r)

	body, header := sendReqToUpstream(req)
	body = replaceURLInResp(body, header)

	w.Write(body)
}

func main() {
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
