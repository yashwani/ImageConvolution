package implementations

//Run the correct version based on the Mode field of the configuration value
func Schedule(config Config) {
	if config.Mode == "s" {
		RunSequential(config)
	} else if config.Mode == "pipeline" {
		RunPipeline(config)
	} else if config.Mode == "bsp" {
		ctx := NewBSPContext(config)
		var idx int
		for idx = 0; idx < config.ThreadCount-1; idx++ {
			go RunBSPWorker(idx, ctx)
		}
		RunBSPWorker(idx, ctx)
	} else if config.Mode == "ws" {
		RunWS(config)
	} else {
		panic("Invalid scheduling scheme given.")
	}
}
