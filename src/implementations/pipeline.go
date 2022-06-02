package implementations

import (
	"encoding/json"
	"fmt"
	"os"
	"src/png"
	"strings"
)

//ImageTaskGenerator reads from effects.txt file and places tasks into a taskGeneratorStream channel
func ImageTaskGenerator(config Config, done chan interface{}) <-chan *png.ImageTask {
	taskGeneratorStream := make(chan *png.ImageTask)
	go func() {
		dirs := strings.Split(config.DataDirs, "+")
		for _, dir := range dirs {

			effectsPathFile := fmt.Sprintf("../data/effects.txt")
			effectsFile, _ := os.Open(effectsPathFile)
			reader := json.NewDecoder(effectsFile)
			for {
				task, imagePathOut, pngImg := ReadImage(reader, dir)
				if pngImg == nil {
					break
				}

				pngImg.AddOutPath(imagePathOut)
				pngImg.Effects = task.Effects

				select {
				case <-done:
					return
				case taskGeneratorStream <- pngImg:
				}
			}
		}

		close(taskGeneratorStream)

	}()

	return taskGeneratorStream

}

//Worker reads from taskGeneratorStream and handles processing of all effects for that specific image
func Worker(taskGeneratorStream <-chan *png.ImageTask, done chan interface{}, idx int, config Config) <-chan *png.ImageTask {
	completedTaskStream := make(chan *png.ImageTask)
	go func() {
		for {
			img, more := <-taskGeneratorStream
			if !more {
				close(completedTaskStream)
				return
			}

			effects := img.Effects
			for i, effect := range effects { //loop over effects for image task
				bufferChan := make(chan bool, config.ThreadCount) //bufferchan used to wait miniworkers once one effect has been processed
				for i := 0; i < config.ThreadCount; i++ {
					go MiniWorker(img, bufferChan, i, config.ThreadCount, effect)
				}

				for {
					if len(bufferChan) == config.ThreadCount {
						break
					}
				}
				if i < len(img.Effects)-1 {
					img.PipeOutIn()
				}
			}

			select {
			case <-done:
				return
			case completedTaskStream <- img:
			}
		}
	}()
	return completedTaskStream
}

//MiniWorker performs actual processing on subsection of image. Enters a boolean value into bufferChan to indicate
//to calling function that it is done processing
func MiniWorker(img *png.ImageTask, bufferChan chan bool, sectionNum int, totalSections int, effect string) {

	img.ApplyEffectToSubsection(effect, sectionNum, totalSections)
	bufferChan <- true

}

//ResultsAggregator mutiplexes in results from completedTaskStream channels that Workers inserted completed tasks into
//it. Tasks are multiplexed into allCompletedTaskStream, which is then processed using an anonymous lambda function
//which saves the completed images to files
func ResultsAggregator(done chan interface{}, channels ...<-chan *png.ImageTask) chan *png.ImageTask {

	allCompletedTaskStream := make(chan *png.ImageTask)

	multiplex := func(c <-chan *png.ImageTask, bufferChan chan bool) {
		for {
			img, more := <-c
			if !more {
				break
			}
			allCompletedTaskStream <- img
		}
		bufferChan <- true

	}

	bufferChan := make(chan bool, len(channels))

	//process everything in allCompletedTaskStream
	go func() {

		for {
			if len(bufferChan) == len(channels) {
				close(allCompletedTaskStream)
			}
			img, more := <-allCompletedTaskStream
			if !more {
				done <- true
				break
			} else {
				err := img.Save(img.ImagePathOut)
				if err != nil {
					return
				}
			}
		}

	}()

	for _, c := range channels {
		go multiplex(c, bufferChan)
	}

	return allCompletedTaskStream

}

//The top level function in this file that runs the pipeline version of the filtering implementation
func RunPipeline(config Config) {

	done := make(chan interface{})

	taskCh := ImageTaskGenerator(config, done)

	workers := make([]<-chan *png.ImageTask, config.ThreadCount)
	for i := 0; i < config.ThreadCount; i++ { // fan out step
		workers[i] = Worker(taskCh, done, i, config)
	}

	ResultsAggregator(done, workers...)

	<-done

}
