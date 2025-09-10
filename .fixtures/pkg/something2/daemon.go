package something2

import "fmt"

type ServiceDaemon struct {
	name string
}

func (s *ServiceDaemon) Start() error {
	fmt.Printf("ServiceDaemon %s starting\n", s.name)
	return nil
}

func (s *ServiceDaemon) Stop() error {
	fmt.Printf("ServiceDaemon %s stopping\n", s.name)
	return nil
}

func (s *ServiceDaemon) GetName() string {
	return s.name
}