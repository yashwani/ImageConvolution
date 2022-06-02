package implementations

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Task struct {
	InPath  string   `json:"inPath"`
	OutPath string   `json:"outPath"`
	Effects []string `json:"effects"`
}

//RunSeuential runs the sequential implementation of the filtering
func RunSequential(config Config) {
	dirs := strings.Split(config.DataDirs, "+")
	for _, dir := range dirs {

		effectsPathFile := fmt.Sprintf("../data/effects.txt")
		effectsFile, _ := os.Open(effectsPathFile)
		reader := json.NewDecoder(effectsFile)
		for { //loops over tasks
			task, imagePathOut, pngImg := ReadImage(reader, dir)
			if pngImg == nil {
				break
			}

			for i, effect := range task.Effects { //loop over every effect

				pngImg.ApplyEffectToSubsection(effect, 0, 1) //apply filter to entirety of image

				if i < len(task.Effects)-1 {
					pngImg.PipeOutIn()
				}
			}

			//Saves the image to a new file
			err := pngImg.Save(imagePathOut)
			if err != nil { //err when saving
				panic(err)
			}

		}
	}

}
