package implementations

type Config struct {
	DataDirs string //Represents the data directories to use to load the images.
	Mode     string // Represents which scheduler scheme to use
	// If Mode == "s" run the sequential version
	// If Mode == "pipeline" run the pipeline version
	// If Mode == "bsp" run the pipeline version
	// These are the only values for Version
	ThreadCount int // Runs the parallel version of the program with the
	// specified number of threads (i.e., goroutines)
}
