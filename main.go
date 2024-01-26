package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)
type ServerI interface{
	getAddress()
	isAlive()
	Serve(rw http.ResponseWriter ,r *http.Request)
}
func(se *Servers) getAddress()string{
	return se.addr
}

func(se *Servers) isAlive() bool{
	return true
}

func(se *Servers) Serve(rw http.ResponseWriter ,r *http.Request){
	se.revProxy.ServeHTTP(rw,r)
}

type Servers struct{
	addr string
	revProxy *httputil.ReverseProxy
}

type LoadBalancer struct{
	port string
	server []Servers
	RoundRobinCount int
}
func newLoadBalancer(port string, servers []Servers) *LoadBalancer{
	return &LoadBalancer{
		port: port,
		RoundRobinCount: 0,
		server: servers,
	}
}

func newServer(addr string) Servers{
	serveruri , err := url.Parse(addr)
	HandleError(err)

	return Servers{
		addr:addr,
		revProxy: httputil.NewSingleHostReverseProxy(serveruri),
	}
}
func HandleError(err error){
	if err!= nil{
		fmt.Println("Error",err)
		os.Exit(1)
	}
}

func(lb *LoadBalancer) getNextAvailableServer() Servers{
	server := lb.server[lb.RoundRobinCount%len(lb.server)]
	for !server.isAlive(){
		lb.RoundRobinCount++
		server = lb.server[lb.RoundRobinCount%len(lb.server)]
	}
	lb.RoundRobinCount++
	return server
}

func(lb *LoadBalancer) ServeProxy(rw http.ResponseWriter, r *http.Request){
	targetServer :=lb.getNextAvailableServer()
	fmt.Println("Forwarding request to address:",targetServer.getAddress())
	targetServer.Serve(rw,r)
}

func main(){
	servers := []Servers{
		newServer("https://opensource.microsoft.com"),
		newServer("https://www.google.com"),
		newServer("https://www.youtube.com"),
	}
	lb := newLoadBalancer("8080",servers)
	HandleRequest := func(rw http.ResponseWriter, r *http.Request){
		lb.ServeProxy(rw,r)
	}
	http.HandleFunc("/",HandleRequest)
	fmt.Println("Serving Requests at localhost:"+lb.port)
	http.ListenAndServe(":"+lb.port,nil)
}