package biscuitjar

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
)

type Jar struct {
	jar   *cookiejar.Jar
	urls  map[*url.URL]bool
	mutex *sync.Mutex
}

func New(options *cookiejar.Options) (*Jar, error) {
	httpjar, err := cookiejar.New(options)
	if err != nil {
		return nil, err
	}

	jar := Jar{
		jar:   httpjar,
		mutex: &sync.Mutex{},
		urls:  map[*url.URL]bool{},
	}

	return &jar, nil
}

func (jar *Jar) Cookies(url *url.URL) []*http.Cookie {
	return jar.jar.Cookies(url)
}

func (jar *Jar) SetCookies(url *url.URL, cookies []*http.Cookie) {
	jar.mutex.Lock()
	defer jar.mutex.Unlock()

	_, ok := jar.urls[url]
	if !ok {
		jar.urls[url] = true
	}

	jar.jar.SetCookies(url, cookies)
}

func (jar *Jar) CookiesAll() map[*url.URL][]*http.Cookie {
	cookies := map[*url.URL][]*http.Cookie{}
	for url, _ := range jar.urls {
		cookies[url] = jar.jar.Cookies(url)
	}

	return cookies
}

func (jar *Jar) Write(writer io.Writer) error {
	allcookies := jar.CookiesAll()
	rawData := map[string][]http.Cookie{}
	for url, cookies := range allcookies {
		urlcookies := []http.Cookie{}
		for _, cookie := range cookies {
			urlcookies = append(urlcookies, *cookie)
		}
		rawData[url.String()] = urlcookies
	}

	return json.NewEncoder(writer).Encode(rawData)
}

func (jar *Jar) Read(reader io.Reader) error {
	allcookies := map[string][]http.Cookie{}

	err := json.NewDecoder(reader).Decode(&allcookies)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	for urlvalue, urlcookies := range allcookies {
		url, err := url.Parse(urlvalue)
		if err != nil {
			return err
		}

		cookies := []*http.Cookie{}
		for _, cookie := range urlcookies {
			cookies = append(cookies, &cookie)
		}

		jar.SetCookies(url, cookies)
	}

	return nil
}
