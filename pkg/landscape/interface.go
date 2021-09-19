package landscape

type Generator func(*Config) (*Landscape, error)
