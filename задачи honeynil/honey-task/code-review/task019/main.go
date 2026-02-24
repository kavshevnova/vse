package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

// ТЗ: Система для обработки множества файлов и HTTP-соединений.
// Требуется корректное управление ресурсами и их освобождение.

type LogProcessor struct {
	outputDir string
}

func NewLogProcessor(outputDir string) *LogProcessor {
	return &LogProcessor{outputDir: outputDir}
}

func (p *LogProcessor) ProcessLogFiles(inputDir string) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		inputPath := inputDir + "/" + entry.Name()
		file, err := os.Open(inputPath)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", inputPath, err)
			continue
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", inputPath, err)
			continue
		}

		processed := p.processContent(content)

		outputPath := p.outputDir + "/processed_" + entry.Name()
		outputFile, err := os.Create(outputPath)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			continue
		}
		defer outputFile.Close()

		_, err = outputFile.Write(processed)
		if err != nil {
			fmt.Printf("Error writing output: %v\n", err)
			continue
		}

		fmt.Printf("Processed: %s -> %s\n", inputPath, outputPath)
	}

	return nil
}

func (p *LogProcessor) processContent(content []byte) []byte {
	timestamp := time.Now().Format(time.RFC3339)
	header := []byte(fmt.Sprintf("Processed at: %s\n\n", timestamp))
	return append(header, content...)
}

func MergeLogFiles(inputFiles []string, outputPath string) error {
	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	for i, inputPath := range inputFiles {
		file, err := os.Open(inputPath)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", inputPath, err)
			continue
		}
		defer file.Close()

		if i > 0 {
			separator := []byte("\n" + "=".repeat(50) + "\n")
			output.Write(separator)
		}

		_, err = io.Copy(output, file)
		if err != nil {
			fmt.Printf("Error copying from %s: %v\n", inputPath, err)
			continue
		}
	}

	return nil
}

func AnalyzeFiles(filePaths []string) (map[string]int64, error) {
	stats := make(map[string]int64)

	for _, path := range filePaths {
		file, err := os.Open(path)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", path, err)
			continue
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			fmt.Printf("Error getting stats for %s: %v\n", path, err)
			continue
		}

		stats[path] = fileInfo.Size()
	}

	return stats, nil
}

func BatchProcessRecords(records []string, batchSize int) error {
	batches := len(records) / batchSize
	if len(records)%batchSize != 0 {
		batches++
	}

	for i := 0; i < batches; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[start:end]

		filename := fmt.Sprintf("batch_%d.txt", i)
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		for _, record := range batch {
			_, err := file.WriteString(record + "\n")
			if err != nil {
				return err
			}
		}

		fmt.Printf("Batch %d written to %s\n", i, filename)
	}

	return nil
}

func CopyDirectory(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dstDir, 0755)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := srcDir + "/" + entry.Name()
		dstPath := dstDir + "/" + entry.Name()

		srcFile, err := os.Open(srcPath)
		if err != nil {
			fmt.Printf("Error opening source %s: %v\n", srcPath, err)
			continue
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			fmt.Printf("Error creating destination %s: %v\n", dstPath, err)
			continue
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			fmt.Printf("Error copying %s to %s: %v\n", srcPath, dstPath, err)
			continue
		}

		fmt.Printf("Copied: %s -> %s\n", srcPath, dstPath)
	}

	return nil
}

func main() {
	testDir := "/tmp/test_logs"
	outputDir := "/tmp/processed_logs"

	os.MkdirAll(testDir, 0755)
	os.MkdirAll(outputDir, 0755)

	for i := 0; i < 100; i++ {
		filename := fmt.Sprintf("%s/log_%d.txt", testDir, i)
		content := fmt.Sprintf("Log entry %d\nTimestamp: %v\n", i, time.Now())
		os.WriteFile(filename, []byte(content), 0644)
	}

	processor := NewLogProcessor(outputDir)
	err := processor.ProcessLogFiles(testDir)
	if err != nil {
		fmt.Printf("Processing error: %v\n", err)
	}

	var filePaths []string
	for i := 0; i < 100; i++ {
		filePaths = append(filePaths, fmt.Sprintf("%s/log_%d.txt", testDir, i))
	}

	stats, err := AnalyzeFiles(filePaths)
	if err != nil {
		fmt.Printf("Analysis error: %v\n", err)
	}
	fmt.Printf("Analyzed %d files\n", len(stats))

	records := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		records[i] = fmt.Sprintf("Record %d", i)
	}

	err = BatchProcessRecords(records, 50)
	if err != nil {
		fmt.Printf("Batch processing error: %v\n", err)
	}

	fmt.Println("All operations completed!")
}
