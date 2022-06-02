package implementations

import (
	"encoding/json"
	"fmt"
	"os"
	"src/png"
	"strings"
	"sync"
)

type SharedContext struct {
	mutex *sync.Mutex
	count int
}

//RunWS runs the parallel workstealing implementation
func RunWS(config Config) {

	stealingScheduler := NewStealingScheduler(config.ThreadCount)
	stealingScheduler.Run()

	effectsPathFile := fmt.Sprintf("../data/effects.txt")
	effectsFile, _ := os.Open(effectsPathFile)
	reader := json.NewDecoder(effectsFile)
	dirs := strings.Split(config.DataDirs, "+")
	for _, dir := range dirs {
		for {
			task, imagePathOut, pngImg := ReadImage(reader, dir)
			if pngImg == nil {
				break
			}

			pngImg.AddOutPath(imagePathOut)
			pngImg.Effects = task.Effects
			stealingScheduler.PushTask(processImage, pngImg)
		}
	}

	stealingScheduler.Done()
	stealingScheduler.Wait()

}

//processImage servers as the runnable function in each task that is inside local queues
func processImage(arg interface{}) {
	pngImg := arg.(*png.ImageTask)
	for i, effect := range pngImg.Effects { //loop over every effect

		pngImg.ApplyEffectToSubsection(effect, 0, 1) //apply filter to entirety of image

		if i < len(pngImg.Effects)-1 {
			pngImg.PipeOutIn()
		}
	}

	err := pngImg.Save(pngImg.ImagePathOut)
	if err != nil {
		return
	}
}
