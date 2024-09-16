package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/go-logr/logr"
)

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

func (ip *ImageProcessor) ProcessFilesAndDirectories(targets []string) error {
	for _, target := range targets {
		err := ip.processFileOrDirectory(target)
		if err != nil {
			ip.logger.Error(err, "Error processing target", "target", target)
		}
	}

	ip.printAggregateStats()
	return nil
}

func (ip *ImageProcessor) processFileOrDirectory(target string) error {
	fileInfo, err := os.Stat(target)
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		return ip.processDirectory(target)
	}
	return ip.processFile(target)
}

func (ip *ImageProcessor) processDirectory(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return ip.processFile(path)
		}
		return nil
	})
}

func (ip *ImageProcessor) processFile(filePath string) error {
	ip.logger.Info("Shrinking image", "file", filePath)
	return ip.shrinkAndUpdateStats(filePath)
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
