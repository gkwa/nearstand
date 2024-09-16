package core

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type ImageShrinker interface {
	Shrink(input string) (string, error)
	GetSizes(input, output string) (int64, int64, error)
	PrintStats(w io.Writer, input, output string, originalSize, newSize int64)
}

type FileManager interface {
	GetFileSize(path string) (int64, error)
}

type ImageMagickShrinker struct {
	fileManager FileManager
	printer     *message.Printer
}

func NewImageMagickShrinker(fm FileManager) *ImageMagickShrinker {
	return &ImageMagickShrinker{
		fileManager: fm,
		printer:     message.NewPrinter(language.English),
	}
}

func (ims *ImageMagickShrinker) Shrink(input string) (string, error) {
	if err := ims.validateFileFormat(input); err != nil {
		return "", err
	}

	outputPath := ims.generateOutputPath(input)

	cmd := exec.Command(
		"convert", input,
		"-resize", "50%",
		"-quality", "60",
		outputPath,
	)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error executing ImageMagick command: %w", err)
	}

	return outputPath, nil
}

func (ims *ImageMagickShrinker) validateFileFormat(input string) error {
	ext := strings.ToLower(filepath.Ext(input))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		return fmt.Errorf("unsupported file format: %s", ext)
	}
	return nil
}

func (ims *ImageMagickShrinker) GetSizes(input, output string) (int64, int64, error) {
	originalSize, err := ims.fileManager.GetFileSize(input)
	if err != nil {
		return 0, 0, fmt.Errorf("error getting original file size: %w", err)
	}

	newSize, err := ims.fileManager.GetFileSize(output)
	if err != nil {
		return 0, 0, fmt.Errorf("error getting new file size: %w", err)
	}

	return originalSize, newSize, nil
}

func (ims *ImageMagickShrinker) PrintStats(w io.Writer, input, output string, originalSize, newSize int64) {
	fmt.Fprintln(w, "Metric             Before   After    Change")
	fmt.Fprintln(w, "------             ------   -----    ------")
	fmt.Fprintf(w, "%-18s %-8s %-8s %s%s\n", "File Size",
		humanize.Bytes(uint64(originalSize)),
		humanize.Bytes(uint64(newSize)),
		ims.sizeChangeSymbol(originalSize, newSize),
		humanize.Bytes(uint64(abs(originalSize-newSize))))
	fmt.Fprintf(w, "%-18s %-8s %-8s\n", "File Path", input, output)
	fmt.Fprintf(w, "Reduction: %.2f%%\n", (1-float64(newSize)/float64(originalSize))*100)
}

func (ims *ImageMagickShrinker) generateOutputPath(input string) string {
	return filepath.Join(filepath.Dir(input), "shrunk_"+filepath.Base(input))
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
