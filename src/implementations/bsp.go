package implementations

import (
	"encoding/json"
	"fmt"
	"os"
	"src/png"
	"strings"
	"sync"
)

type bspImageTask struct {
	img          *png.ImageTask //The original pixels before applying the effect
	ImagePathOut string
	effect       string
	save         bool
}

type bspWorkerContext struct {
	// Define the necessary fields for your implementation
	bspTasks         []*bspImageTask
	cond             sync.Cond
	finishedLock     *sync.Mutex
	sleepLock        *sync.Mutex
	total_threads    int
	queueIndex       *int
	threads_finished *int
	flag             *bool
}

func convertTobspImageTask(task *png.ImageTask) bspImageTask {
	bspTask := bspImageTask{}
	bspTask.img = task
	bspTask.ImagePathOut = task.ImagePathOut
	return bspTask
}

//FillQueue breaks the input effects.txt file into an ordered list of tasks which can be processed using BSP
func FillQueue(config Config) []*bspImageTask {
	queue := make([]*bspImageTask, 0)

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
			for i, effect := range task.Effects {
				bspTask := convertTobspImageTask(pngImg)
				bspTask.effect = effect
				if i == len(task.Effects)-1 {
					bspTask.save = true
				}
				queue = append(queue, &bspTask)
			}
		}
	}
	return queue

}

//NewBSPContext creates a bsp context that can be used across bsp workers
func NewBSPContext(config Config) *bspWorkerContext {
	bspTasks := FillQueue(config)

	queueIndex := 0
	threads_finished := 0
	sleepLock := sync.Mutex{}
	flag := true
	ctx := bspWorkerContext{bspTasks: bspTasks, cond: sync.Cond{L: &sleepLock}, finishedLock: &sync.Mutex{}, sleepLock: &sleepLock,
		total_threads: config.ThreadCount, queueIndex: &queueIndex, threads_finished: &threads_finished, flag: &flag}
	return &ctx
}

//bspWorker applies the filter to a subsection of the image
func bspWorker(img *png.ImageTask, sectionNum int, totalSections int, effect string) {

	img.ApplyEffectToSubsection(effect, sectionNum, totalSections)
}

//RunBSPWorker completes one task in the queue and waits for all other threads to finish the same task before
//continuing onto the next task
func RunBSPWorker(id int, ctx *bspWorkerContext) {
	for {
		queueIndex := ctx.queueIndex
		imgTask := ctx.bspTasks[*queueIndex]
		img := imgTask.img
		bspWorker(img, id, ctx.total_threads, imgTask.effect)

		ctx.sleepLock.Lock()

		*ctx.threads_finished += 1
		if *ctx.threads_finished < ctx.total_threads {
			for *ctx.flag {
				ctx.cond.Wait()
			}

		} else {
			if imgTask.save {
				err := img.Save(imgTask.ImagePathOut)
				if err != nil {
					return
				}
			}
			*ctx.queueIndex += 1
			*ctx.flag = false
			img.PipeOutIn()
			ctx.cond.Broadcast()

		}
		*ctx.threads_finished -= 1
		if *ctx.threads_finished == 0 {
			*ctx.flag = true
		}
		ctx.sleepLock.Unlock()

		if *ctx.queueIndex == len(ctx.bspTasks) {
			return
		}

	}
}
