package main

import (
	"io"
	"log"
	"net"
	"net/http"
)

func main() {
	http.ListenAndServe(":8080", new(Proxy))
}

type Proxy struct{}

func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s %s\n", req.Method, req.Host, req.RemoteAddr)
	if req.Method != "CONNECT" {
		p.HTTP(rw, req)
	} else {
		p.HTTPS(rw, req)
	}

}

func (p *Proxy) HTTP(rsp http.ResponseWriter, req *http.Request) {
	outReq := new(http.Request)
	*outReq = *req

	res, err := http.DefaultTransport.RoundTrip(outReq)
	if err != nil {
		rsp.WriteHeader(http.StatusBadGateway)
		rsp.Write([]byte(err.Error()))
		return
	}
	defer res.Body.Close()

	for key, value := range res.Header {
		for _, v := range value {
			rsp.Header().Add(key, v)
		}
	}

	rsp.WriteHeader(res.StatusCode)
	io.Copy(rsp, res.Body)
}

func (p *Proxy) HTTPS(rw http.ResponseWriter, req *http.Request) {
	host := req.URL.Host
	hij, ok := rw.(http.Hijacker)
	if !ok {
		log.Printf("HTTP Server does not support hijacking")
	}

	client, _, err := hij.Hijack()
	if err != nil {
		return
	}

	server, err := net.Dial("tcp", host)
	if err != nil {
		return
	}
	client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))

	go io.Copy(server, client)
	go io.Copy(client, server)
}
