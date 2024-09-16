package core

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/dustin/go-humanize"
)

type ImageShrinker interface {
	ShrinkAndGetSizes(input string) (int64, int64, error)
}

type FileManager interface {
	GetFileSize(path string) (int64, error)
}

type ImageMagickShrinker struct {
	fileManager FileManager
}

func NewImageMagickShrinker(fm FileManager) *ImageMagickShrinker {
	return &ImageMagickShrinker{fileManager: fm}
}

func (ims *ImageMagickShrinker) ShrinkAndGetSizes(input string) (int64, int64, error) {
	ext := filepath.Ext(input)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		return 0, 0, fmt.Errorf("unsupported file format: %s", ext)
	}

	originalSize, err := ims.fileManager.GetFileSize(input)
	if err != nil {
		return 0, 0, fmt.Errorf("error getting original file size: %w", err)
	}

	outputPath := filepath.Join(filepath.Dir(input), "shrunk_"+filepath.Base(input))

	cmd := exec.Command(
		"convert", input,
		"-resize", "50%",
		"-quality", "60",
		outputPath,
	)
	err = cmd.Run()
	if err != nil {
		return 0, 0, fmt.Errorf("error executing ImageMagick command: %w", err)
	}

	tileCmd := exec.Command(
		"convert", outputPath,
		"-crop", "1024x768",
		filepath.Join(filepath.Dir(input), "tile_%d.jpg"),
	)
	err = tileCmd.Run()
	if err != nil {
		return 0, 0, fmt.Errorf("error splitting shrunk image into tiles: %w", err)
	}

	newSize, err := ims.fileManager.GetFileSize(outputPath)
	if err != nil {
		return 0, 0, fmt.Errorf("error getting new file size: %w", err)
	}

	ims.printStats(input, outputPath, originalSize, newSize)

	return originalSize, newSize, nil
}

func (ims *ImageMagickShrinker) printStats(input, outputPath string, originalSize, newSize int64) {
	fmt.Println("Metric             Before   After    Change")
	fmt.Println("------             ------   -----    ------")
	fmt.Printf("%-18s %-8s %-8s %s%s\n", "File Size",
		humanize.Bytes(uint64(originalSize)),
		humanize.Bytes(uint64(newSize)),
		ims.sizeChangeSymbol(originalSize, newSize),
		humanize.Bytes(uint64(abs(originalSize-newSize))))
	fmt.Printf("%-18s %-8s %-8s\n", "File Path", input, outputPath)
	fmt.Printf("Reduction: %.2f%%\n", (1-float64(newSize)/float64(originalSize))*100)
}

func (ims *ImageMagickShrinker) sizeChangeSymbol(originalSize, newSize int64) string {
	if newSize > originalSize {
		return "+"
	}
	return "-"
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
