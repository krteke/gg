package internal

type ScanConfig struct {
	Root   string
	Output string
	Config string
}

func Scan(config ScanConfig) error {
	return nil
}
