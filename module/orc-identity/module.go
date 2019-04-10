package identity

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orc"
)

var (
	identity *IdentityClaim

	stillMutable bool = true
	mu           sync.Mutex
)

func Identity() IdentityClaim {
	if stillMutable {
		logrus.Warningf("Warning: Identity requested while still mutable")
	}
	return *identity
}

func MutateIdentity(mutator func(*IdentityClaim)) {
	logrus.Infof("Mutating identity: before=%v", identity)
	defer func() { logrus.Infof("Mutating identity: after=%v", identity) }()
	mu.Lock()
	defer mu.Unlock()

	if identity == nil {
		logrus.Fatalf("Error: too early to mutate Identity")
	}

	if !stillMutable {
		logrus.Fatalf("Error: too late to mutate Identity")
	}

	mutator(identity)
}

type Module struct {
	identity *IdentityClaim
}

func (m *Module) ModuleName() string { return "Identity" }

func New(cfg Config) (*Module, error) {
	mu.Lock()
	defer mu.Unlock()

	if cfg.ProgramName == "" {
		return nil, fmt.Errorf("No program name provided")
	}
	if identity != nil {
		return nil, fmt.Errorf("Identity was already set")
	}

	id, err := cfg.create()
	if err != nil {
		return nil, err
	}

	m := &Module{identity}
	identity = id

	return m, nil
}

func MustNew(cfg Config) *Module {
	m, err := New(cfg)
	if err != nil {
		logrus.Fatalf("Failed to create Identity module: %v", err)
	}
	return m
}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnStart(func() error {
		mu.Lock()
		defer mu.Unlock()

		if err := validateIdentity(identity); err != nil {
			return err
		}

		formatted, err := json.Marshal(identity)
		if err != nil {
			return err
		}
		stillMutable = false

		logrus.Infof("Identity finalised: %s", string(formatted))

		return nil
	})
}
