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

// Server Интерфейс для описания сервера
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

// Создание сервера соответствующего интерфейсу Server
func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleError(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

// NewLoadBalancer Создать новый балансировщик
func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:       port,
		roundCount: 0,
		servers:    servers,
	}
}

// Управление ошибками
func handleError(err error) {
	if err != nil {
		fmt.Printf("Ошибка %v\n", err)
		os.Exit(1)
	}
}

// Address Получение адреса сервера
func (s *simpleServer) Address() string { return s.addr }

// IsAlive Узнать жив ли сервер
func (s *simpleServer) IsAlive() bool { return true }

func (s *simpleServer) Server(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

// Получение доступных серверов
func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundCount++
		server = lb.servers[lb.roundCount%len(lb.servers)]
	}
	lb.roundCount++
	return server
}

// Запуск сервера через прокси
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
	} //Добавление списка серверов
	lb := NewLoadBalancer("8000", servers) //Создание нового балансировщика
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		lb.serveProxy(w, r)
	} //Редирект когда набирается адрес
	http.HandleFunc("/", handleRedirect) //Привязка редиректа
	fmt.Printf("Обслуживаение реквестов на порту %v\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil) //запуск прослушки сервера по порту
}
