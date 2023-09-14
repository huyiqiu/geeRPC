package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect SelectMode = iota
	RoundRobinSelect
)

type Discovery interface {
	Refresh() error
	Update(server []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

type MultiServerDiscovery struct {
	r       *rand.Rand
	mu      sync.RWMutex
	servers []string
	index   int
}

func NewMultiServerDiscovery(servers []string) *MultiServerDiscovery {
	d := &MultiServerDiscovery{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
		servers: servers,
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

var _ Discovery = (*MultiServerDiscovery)(nil)

func (d *MultiServerDiscovery) Refresh() error {
	return nil
}

func (d *MultiServerDiscovery) Update(server []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = server
	return nil
}

func (d *MultiServerDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelect:
		s := d.servers[(d.index) % n]
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

func (d *MultiServerDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	servers := make([]string, len(d.servers))
	copy(servers, d.servers)
	return d.servers, nil
}