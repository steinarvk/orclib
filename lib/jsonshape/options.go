package jsonshape

type options struct {
	keptFields map[string]bool
}

func buildOptions(opts []Option) (options, error) {
	rv := options{keptFields: map[string]bool{}}
	for _, opt := range opts {
		if opt != nil {
			if err := opt(&rv); err != nil {
				return options{}, nil
			}
		}
	}
	return rv, nil
}

type Option func(*options) error

func WithPreservedFields(fieldName ...string) Option {
	return func(opts *options) error {
		for _, n := range fieldName {
			opts.keptFields[n] = true
		}
		return nil
	}
}
