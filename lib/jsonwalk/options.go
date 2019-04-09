package jsonwalk

import "fmt"

type options struct {
	path              string
	elideArrayIndices bool
}

type Option func(*options) error

func WithRootPath(s string) Option {
	return func(opts *options) error {
		if opts.path != "" {
			return fmt.Errorf("Root path already set: would set %q, already %q", s, opts.path)
		}
		opts.path = s
		return nil
	}
}

func WithoutArrayIndices() Option {
	return func(opts *options) error {
		opts.elideArrayIndices = true
		return nil
	}
}
