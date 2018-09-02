package wsnet

// Options is the server config
type Options struct {
	ReadBufferSize  int
	WriteBufferSize int
	ReadDeadline    int
	WriteDeadline   int
	OutgoinSize     int
}

// NewOptions returns default options
func NewOptions() *Options {
	return defaultOptions()
}

func defaultOptions() *Options {
	return &Options{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		ReadDeadline:    5000,
		WriteDeadline:   5000,
		OutgoinSize:     20,
	}
}
