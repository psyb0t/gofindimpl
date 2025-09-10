package something3

import "fmt"

type MicroService struct {
	id   string
	host string
}

func (m *MicroService) Start() error {
	fmt.Printf("MicroService %s starting on %s\n", m.id, m.host)
	return nil
}

func (m *MicroService) Stop() error {
	fmt.Printf("MicroService %s stopping\n", m.id)
	return nil
}

func (m *MicroService) GetName() string {
	return fmt.Sprintf("MicroService-%s", m.id)
}