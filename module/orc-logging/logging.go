package logging

import (
	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orc"
)

type Module struct{}

var M = &Module{}

func (m *Module) ModuleName() string { return "Logging" }

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	var logLevel string
	hooks.OnUse(func(ctx orc.UseContext) {
		ctx.Flags.StringVar(&logLevel, "log_level", "info", "logging level")
	})

	hooks.OnValidate(func() error {
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			return err
		}

		logrus.AddHook(logStatHook{})

		// Hack: mutating in Validate, just to set this as soon as possible.
		logrus.SetLevel(level)

		return nil
	})
}
