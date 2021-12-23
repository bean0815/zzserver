package zztools

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

//httpPostForm resp, err := http.PostForm("http://www.01happy.com/demo/accept.php",
// url.Values{"key": {"Value"}, "id": {"123"}})
func HttpPostForm(url string, data url.Values) (body []byte, err error) {
	resp, err := http.PostForm(url, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	return
}

//http.Post("http://www.01happy.com/demo/accept.php","application/x-www-form-urlencoded", strings.NewReader("name=cjb"))
func HttpPost(url, data string) (body []byte, err error) {
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	return
}

//发送http Get 请求
func HttpGet(url string) (body []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	return
}
