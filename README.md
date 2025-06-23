# Балансировщик запросов

Работает так:
1. Вводится адрес в браузере "http://localhost:8000"
2. Сервер персылает на доступные адреса из массива
```go
	servers := []Server{
		newSimpleServer("https://yandex.ru"),
		newSimpleServer("https://mail.ru"),
		newSimpleServer("https://github.com"),
	} 
```