package cfg

// Resolver is an interface for resolving special strings in a config file to a concrete value.
type Resolver interface {
	Resolve(string) (string, error)
}

// Resolvers is a slice of Resolver
type Resolvers []Resolver

func (r Resolvers) Resolve(in string) (string, error) {
	for _, resolver := range r {
		var err error

		in, err = resolver.Resolve(in)
		if err != nil {
			return "", err
		}
	}

	return in, nil
}
