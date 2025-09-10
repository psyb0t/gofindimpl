package something1

import "fmt"

type WebServer struct {
	port int
}

func (w *WebServer) Start() error {
	fmt.Printf("WebServer starting on port %d\n", w.port)
	return nil
}

func (w *WebServer) Stop() error {
	fmt.Println("WebServer stopping")
	return nil
}

func (w *WebServer) GetName() string {
	return "WebServer"
}