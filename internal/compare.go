package internal

type CompareConfig struct {
	Job    string
	Source string
	Target string
	Output string
}

func (c *CompareConfig) Compare() error {

	return nil
}
