package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

// Базовый сервер
type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy // Входящий запрос направляет на другие сервера
}

type Server interface {
	Address() string
	IsAlive() bool
	Server(w http.ResponseWriter, r *http.Request)
}

type LoadBalancer struct {
	port       string
	roundCount int
	servers    []Server
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleError(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:       port,
		roundCount: 0,
		servers:    servers,
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("Ошибка %v\n", err)
		os.Exit(1)
	}
}

func (s *simpleServer) Address() string { return s.addr }

func (s *simpleServer) IsAlive() bool { return true }

func (s *simpleServer) Server(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundCount++
		server = lb.servers[lb.roundCount%len(lb.servers)]
	}
	lb.roundCount++
	return server
}

func (lb *LoadBalancer) serveProxy(w http.ResponseWriter, r *http.Request) {
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("Отправка на адрес %v\n", targetServer.Address())
	targetServer.Server(w, r)
}

func main() {
	servers := []Server{
		newSimpleServer("https://ya.ru"),
		newSimpleServer("https://mail.ru"),
		newSimpleServer("https://github.com"),
	}
	lb := NewLoadBalancer("8000", servers)
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		lb.serveProxy(w, r)
	}
	http.HandleFunc("/", handleRedirect)
	fmt.Printf("Обслуживаение реквестов на порту %v\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
