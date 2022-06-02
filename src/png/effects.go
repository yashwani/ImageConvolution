// Package png allows for loading png images and applying
// image flitering effects on them.
package png

import "image/color"

// Grayscale applies a grayscale filtering effect a subsection of the image. The subsection is calculated by
//dividing (relatively) evenly the image int totalSections sections, and choosing the sectionNum
//(indexed starting at 0) section
func (img *ImageTask) Grayscale(sectionNum int, totalSections int) {

	// Bounds returns defines the dimensions of the image. Always
	// use the bounds Min and Max fields to get Out the width
	// and height for the image
	bounds := img.Out.Bounds()

	sectionRange := bounds.Dy() / totalSections //integer division
	yTopSection := sectionNum * sectionRange
	yBotSection := (sectionNum + 1) * sectionRange
	if sectionNum == totalSections-1 { //if last sectionNum, set high to bottom of image
		yBotSection = bounds.Max.Y
	}

	for y := yTopSection; y < yBotSection; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			//Returns the pixel (i.e., RGBA) value at a (x,y) position
			// Note: These get returned as int32 so based on the math you'll
			// be performing you'll need to do a conversion to float64(..)
			r, g, b, a := img.In.At(x, y).RGBA()

			//Note: The values for r,g,b,a for this assignment will range between [0, 65535].
			//For certain computations (i.e., convolution) the values might fall outside this
			// range so you need to clamp them between those values.
			greyC := clamp(float64(r+g+b) / 3)

			//Note: The values need to be stored back as uint16 (I know weird..but there's valid reasons
			// for this that I won't get into right now).
			img.Out.Set(x, y, color.RGBA64{greyC, greyC, greyC, uint16(a)})
		}
	}
}

//filter applies a 3*3 filter to each color channel of an rgb pixel and returns the filtered pixel. handles padding
func (img *ImageTask) filter(x int, y int, kernel [9]float64, xLeft int, xRight int, yTop int, yBot int, isTop bool, isBot bool) (uint16, uint16, uint16, uint16) {

	// 2d positions:
	// 0 1 2
	// 3 4 5
	// 6 7 8

	xPosKernel := [9]int{x - 1, x, x + 1, x - 1, x, x + 1, x - 1, x, x + 1}
	yPosKernel := [9]int{y - 1, y - 1, y - 1, y, y, y, y + 1, y + 1, y + 1}

	sum_r := float64(0)
	sum_g := float64(0)
	sum_b := float64(0)

	for i := 0; i < len(kernel); i++ {
		xPos := xPosKernel[i]
		yPos := yPosKernel[i]
		kernelMultiplier := kernel[i]

		var r, g, b, _ uint32
		//pad during loop
		if xPos < xLeft || xPos == xRight || (isTop && yPos < yTop) || (isBot && yPos == yBot) {
			r, g, b, _ = 0, 0, 0, 0
		} else {
			r, g, b, _ = img.In.At(xPos, yPos).RGBA()
		}
		sum_r += float64(r) * kernelMultiplier
		sum_g += float64(g) * kernelMultiplier
		sum_b += float64(b) * kernelMultiplier

	}

	_, _, _, a := img.In.At(x, y).RGBA()

	return clamp(sum_r), clamp(sum_g), clamp(sum_b), uint16(a)

}

//Convolve convolves a specific section of an image with a kernel of size 3*3/ The subsection is calculated by
//dividing (relatively) evenly the image int totalSections sections, and choosing the sectionNum
//(indexed starting at 0) section
func (img *ImageTask) Convolve(kernel [9]float64, sectionNum int, totalSections int) {
	//sections are divide horizontally
	//if first section or last section, pads top or bottom
	//for all sections, pad left and right

	//convolving a kernel with the entire image uses parameters: sectionNum = 0, totalSections = 1

	bounds := img.Out.Bounds()

	xLeft := bounds.Min.X
	xRight := bounds.Max.X
	yTop := bounds.Min.Y
	yBot := bounds.Max.Y

	sectionRange := bounds.Dy() / totalSections //integer division
	yTopSection := sectionNum * sectionRange
	yBotSection := (sectionNum + 1) * sectionRange
	if sectionNum == totalSections-1 { //if last sectionNum, set high to bottom of image
		yBotSection = bounds.Max.Y
	}

	for y := yTopSection; y < yBotSection; y++ {
		for x := xLeft; x < xRight; x++ {

			r, g, b, a := img.filter(x, y, kernel, xLeft, xRight, yTop, yBot, sectionNum == 0, sectionNum == totalSections-1)
			img.Out.Set(x, y, color.RGBA64{R: r, G: g, B: b, A: a})

		}
	}

}

//Applies a filter to a specific subsection of the image
//Filters: sharpen, edge detection, blur, grayscale
func (pngImg *ImageTask) ApplyEffectToSubsection(effect string, sectionNum int, totalSections int) {
	switch {
	case effect == "S":
		sharpenKernel := [9]float64{0, -1, 0, -1, 5, -1, 0, -1, 0}
		pngImg.Convolve(sharpenKernel, sectionNum, totalSections)
	case effect == "E":
		edgeDetectionKernel := [9]float64{-1, -1, -1, -1, 8, -1, -1, -1, -1}
		pngImg.Convolve(edgeDetectionKernel, sectionNum, totalSections)
	case effect == "B":
		blurKernel := [9]float64{1 / 9.0, 1 / 9, 1 / 9.0, 1 / 9.0, 1 / 9.0, 1 / 9.0, 1 / 9.0, 1 / 9.0, 1 / 9.0}
		pngImg.Convolve(blurKernel, sectionNum, totalSections)
	case effect == "G":
		pngImg.Grayscale(sectionNum, totalSections)
	}
}
