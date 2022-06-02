package implementations

import (
	"encoding/json"
	"path"
	"src/png"
	"strings"
)

//a helper function to read in an image from a json Decoder and give back the task, the output path, and the image itself
func ReadImage(reader *json.Decoder, dir string) (Task, string, *png.ImageTask) {
	var task Task
	if err := reader.Decode(&task); err != nil {
		return task, "", nil //returns when no more lines left
	}

	//create image in and out paths
	imagePathIn := path.Join("../data", "in", dir, task.InPath)
	imagePathOut := path.Join("../data", "out", dir+"_"+strings.Split(task.InPath, ".")[0]+"_Out"+".png")

	//Loads the png image and returns the image or an error
	pngImg, err := png.Load(imagePathIn)
	if err != nil { //err when loading
		panic(err)
	}
	return task, imagePathOut, pngImg
}
