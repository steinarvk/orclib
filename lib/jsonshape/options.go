package jsonshape

type options struct {
}

func buildOptions(opts []Option) (options, error) {
	rv := options{}
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
