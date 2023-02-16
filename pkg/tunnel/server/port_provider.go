package server

import (
	"sync"

	"emperror.dev/errors"
)

var ErrInvalidPortRange = errors.New("invalid port range")

type portProvider struct {
	portMin int
	portMax int
	ports   map[int]bool

	mu sync.Mutex
}

type PortProvider interface {
	GetFreePort() int
	ReleasePort(int)
}

func NewPortProvider(portMin, portMax int) (PortProvider, error) {
	if portMin > portMax {
		return nil, errors.WithStackIf(ErrInvalidPortRange)
	}

	return &portProvider{
		ports:   make(map[int]bool),
		portMin: portMin,
		portMax: portMax,
		mu:      sync.Mutex{},
	}, nil
}

func (p *portProvider) GetFreePort() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := p.portMin; i < p.portMax; i++ {
		if _, ok := p.ports[i]; !ok {
			p.ports[i] = true
			return i
		}
	}

	return 0
}

func (p *portProvider) ReleasePort(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.ports, port)
}
