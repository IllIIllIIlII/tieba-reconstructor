package network

import (
	"net/http"
	"net/url"
)

const defaultRetryNum int = 3

func Get(urli string) (*http.Response, error) {
	client := &http.Client{}
	header := &http.Header{}
	header.Add("User-Agent", "Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; InfoPath.3)")
	req := &http.Request{
		Method: "GET",
	}
	req.URL, _ = url.Parse(urli)
	req.Header = *header
	res, err := client.Do(req)
	count := 1
	for err != nil && count < defaultRetryNum {
		res, err = client.Do(req)
	}
	return res, err
}

func Post(urli string) (*http.Response, error) {
	client := &http.Client{}
	header := &http.Header{}
	header.Add("User-Agent", "Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; InfoPath.3)")
	req := &http.Request{
		Method: "POST",
	}
	req.URL, _ = url.Parse(urli)
	req.Header = *header
	res, err := client.Do(req)
	count := 1
	for err != nil && count < defaultRetryNum {
		res, err = client.Do(req)
	}
	return res, err
}
func GetRequestFromUrl(s string) (*http.Request, error) {
	URL, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	header := http.Header{}
	header.Add("User-Agent", "Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; InfoPath.3)")
	r := &http.Request{
		URL:    URL,
		Header: header,
	}
	return r, nil
}

func Retry(client http.Client, req *http.Request, n int) (*http.Response, error) {
	var i int
	var err error = nil
	var resp *http.Response
	for i = 0; i < n; i++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}
	}
	return resp, err
}
