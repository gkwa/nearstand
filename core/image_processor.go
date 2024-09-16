package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/go-logr/logr"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type ImageProcessor struct {
	shrinker   ImageShrinker
	logger     logr.Logger
	totalStats TotalStats
	statsMutex sync.Mutex
	writer     io.Writer
	printer    *message.Printer
}

type TotalStats struct {
	OriginalSize int64
	ShrunkSize   int64
	FileCount    int
	SkippedCount int
}

func NewImageProcessor(shrinker ImageShrinker, logger logr.Logger, writer io.Writer) *ImageProcessor {
	return &ImageProcessor{
		shrinker: shrinker,
		logger:   logger,
		writer:   writer,
		printer:  message.NewPrinter(language.English),
	}
}

func (ip *ImageProcessor) ProcessFilesAndDirectories(targets []string, reshrink bool) error {
	for _, target := range targets {
		err := ip.processFileOrDirectory(target, reshrink)
		if err != nil {
			ip.logger.Error(err, "Error processing target", "target", target)
		}
	}

	ip.printAggregateStats()
	return nil
}

func (ip *ImageProcessor) processFileOrDirectory(target string, reshrink bool) error {
	fileInfo, err := os.Stat(target)
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		return ip.processDirectory(target, reshrink)
	}
	return ip.processFile(target, reshrink)
}

func (ip *ImageProcessor) processDirectory(dirPath string, reshrink bool) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			ip.logger.Error(err, "Error accessing file", "path", path)
			return nil
		}
		if !info.IsDir() {
			err := ip.processFile(path, reshrink)
			if err != nil {
				ip.logger.Error(err, "Error processing file", "path", path)
			}
		}
		return nil
	})
}

func (ip *ImageProcessor) processFile(filePath string, reshrink bool) error {
	if !reshrink && strings.HasPrefix(filepath.Base(filePath), "shrunk_") {
		ip.logger.Info("Skipping already shrunk image", "file", filePath)
		ip.incrementSkippedCount()
		return nil
	}

	ip.logger.Info("Attempting to shrink image", "file", filePath)
	outputPath, err := ip.shrinker.Shrink(filePath)
	if err != nil {
		if strings.Contains(err.Error(), "unsupported file format") {
			ip.logger.Info("Skipping unsupported file", "file", filePath)
			ip.incrementSkippedCount()
			return nil
		}
		return err
	}

	originalSize, newSize, err := ip.shrinker.GetSizes(filePath, outputPath)
	if err != nil {
		return err
	}

	ip.shrinker.PrintStats(ip.writer, filePath, outputPath, originalSize, newSize)

	return ip.updateStats(originalSize, newSize)
}

func (ip *ImageProcessor) updateStats(originalSize, newSize int64) error {
	ip.statsMutex.Lock()
	defer ip.statsMutex.Unlock()

	ip.totalStats.OriginalSize += originalSize
	ip.totalStats.ShrunkSize += newSize
	ip.totalStats.FileCount++

	return nil
}

func (ip *ImageProcessor) incrementSkippedCount() {
	ip.statsMutex.Lock()
	defer ip.statsMutex.Unlock()

	ip.totalStats.SkippedCount++
}

func (ip *ImageProcessor) printAggregateStats() {
	fmt.Fprintln(ip.writer, "\nAggregate Statistics:")
	fmt.Fprintf(ip.writer, "Total files processed: %s\n", ip.printer.Sprintf("%d", ip.totalStats.FileCount))
	fmt.Fprintf(ip.writer, "Total files skipped: %s\n", ip.printer.Sprintf("%d", ip.totalStats.SkippedCount))
	fmt.Fprintf(ip.writer, "Total original size: %s\n", humanize.Bytes(uint64(ip.totalStats.OriginalSize)))
	fmt.Fprintf(ip.writer, "Total reduced size: %s\n", humanize.Bytes(uint64(ip.totalStats.ShrunkSize)))
	reductionSize := ip.totalStats.OriginalSize - ip.totalStats.ShrunkSize
	reductionPercentage := float64(reductionSize) / float64(ip.totalStats.OriginalSize) * 100
	fmt.Fprintf(ip.writer, "Total size reduction: %s (%.2f%%)\n", humanize.Bytes(uint64(reductionSize)), reductionPercentage)
}
