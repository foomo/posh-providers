package sesamy

// Config maps a config set name to the sesamy YAML files that are merged, in
// order, to produce the config each command runs with. The keys are the names
// accepted as the `config` argument.
type Config map[string][]string
