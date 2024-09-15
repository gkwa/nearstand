package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/go-logr/logr"
)

type ImageShrinker interface {
	ShrinkAndGetSizes(input string) (int64, int64, error)
}

type FileManager interface {
	GetFileSize(path string) (int64, error)
}

type DefaultFileManager struct{}

func (dfm DefaultFileManager) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

type ImageMagickShrinker struct {
	fileManager FileManager
}

func NewImageMagickShrinker(fm FileManager) *ImageMagickShrinker {
	return &ImageMagickShrinker{fileManager: fm}
}

type ImageProcessor struct {
	shrinker   ImageShrinker
	logger     logr.Logger
	totalStats TotalStats
	statsMutex sync.Mutex
}

type TotalStats struct {
	OriginalSize int64
	ShrunkSize   int64
	FileCount    int
}

func NewImageProcessor(shrinker ImageShrinker, logger logr.Logger) *ImageProcessor {
	return &ImageProcessor{
		shrinker: shrinker,
		logger:   logger,
	}
}

func (ip *ImageProcessor) Process(target string) error {
	fileInfo, err := os.Stat(target)
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		err = filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				ip.logger.Info("Shrinking image", "file", path)
				return ip.shrinkAndUpdateStats(path)
			}
			return nil
		})
	} else {
		ip.logger.Info("Shrinking image", "file", target)
		err = ip.shrinkAndUpdateStats(target)
	}

	if err != nil {
		return err
	}

	ip.printAggregateStats()
	return nil
}

func (ip *ImageProcessor) shrinkAndUpdateStats(input string) error {
	originalSize, newSize, err := ip.shrinker.ShrinkAndGetSizes(input)
	if err != nil {
		return err
	}

	ip.statsMutex.Lock()
	defer ip.statsMutex.Unlock()

	ip.totalStats.OriginalSize += originalSize
	ip.totalStats.ShrunkSize += newSize
	ip.totalStats.FileCount++

	return nil
}

func (ip *ImageProcessor) printAggregateStats() {
	fmt.Println("\nAggregate Statistics:")
	fmt.Printf("Total files processed: %d\n", ip.totalStats.FileCount)
	fmt.Printf("Total original size: %s\n", humanize.Bytes(uint64(ip.totalStats.OriginalSize)))
	fmt.Printf("Total shrunk size: %s\n", humanize.Bytes(uint64(ip.totalStats.ShrunkSize)))
	reductionSize := ip.totalStats.OriginalSize - ip.totalStats.ShrunkSize
	reductionPercentage := float64(reductionSize) / float64(ip.totalStats.OriginalSize) * 100
	fmt.Printf("Total size reduction: %s (%.2f%%)\n", humanize.Bytes(uint64(reductionSize)), reductionPercentage)
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

	fmt.Println("Metric             Before   After    Change")
	fmt.Println("------             ------   -----    ------")
	fmt.Printf("%-18s %-8s %-8s %s%s\n", "File Size",
		humanize.Bytes(uint64(originalSize)),
		humanize.Bytes(uint64(newSize)),
		ims.sizeChangeSymbol(originalSize, newSize),
		humanize.Bytes(uint64(abs(originalSize-newSize))))
	fmt.Printf("%-18s %-8s %-8s\n", "File Path", input, outputPath)
	fmt.Printf("Reduction: %.2f%%\n", (1-float64(newSize)/float64(originalSize))*100)

	return originalSize, newSize, nil
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

