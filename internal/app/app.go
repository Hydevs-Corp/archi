package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"archi/internal/analyzer"
	"archi/internal/config"
)

type App struct {
	config   *config.Config
	analyzer *analyzer.Analyzer
}

func New(cfg *config.Config) *App {
	return &App{
		config:   cfg,
		analyzer: analyzer.New(cfg),
	}
}

func (a *App) PerformCountAnalysis(targetDir string) error {
	fmt.Println("🧮 Estimate mode: Analyzing directory structure...")

	estimation, err := a.analyzer.PerformCountAnalysis(targetDir)
	if err != nil {
		return fmt.Errorf("error performing count analysis: %w", err)
	}

	estimationContent := analyzer.GenerateEstimationMarkdown(estimation)
	estimationFile := filepath.Join(a.config.DefaultOutputDir, a.config.EstimationFile)
	err = os.WriteFile(estimationFile, []byte(estimationContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing estimation file: %w", err)
	}

	fmt.Printf("\n📊 Estimation Complete!\n")
	fmt.Printf("   Total files: %d\n", estimation.TotalFiles)
	fmt.Printf("   Total folders: %d\n", estimation.TotalFolders)
	fmt.Printf("   Estimated execution time: %s\n", a.formatDuration(estimation.TotalEstimatedTime))
	fmt.Printf("   Estimation saved to: %s\n", estimationFile)

	return nil
}

func (a *App) PerformFullAnalysis(targetDir string) error {
	fmt.Println("🔍 Analyzing directory structure...")

	var totalFiles, totalDirs int
	var extractableFiles, skippedFiles int
	var fileTypes = make(map[string]int)
	var skippedTypes = make(map[string]int)

	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			totalDirs++
		} else {
			totalFiles++
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if a.isExtractableFile(ext) {
				extractableFiles++
				fileTypes[ext]++
			} else if a.isImageFile(ext) {
				skippedFiles++
				skippedTypes[ext]++
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error analyzing directory: %w", err)
	}

	fmt.Printf("\n📊 Directory Analysis Complete:\n")
	fmt.Printf("   Total directories: %d\n", totalDirs)
	fmt.Printf("   Total files: %d\n", totalFiles)
	fmt.Printf("   Files with extractable content: %d\n", extractableFiles)
	fmt.Printf("   Image files (will be analyzed with vision API): %d\n", skippedFiles)

	if len(skippedTypes) > 0 {
		fmt.Printf("\n🖼️  Image file types that will be analyzed with vision API:\n")
		for ext, count := range skippedTypes {
			fmt.Printf("   %s: %d files\n", ext, count)
		}
	}

	fmt.Printf("\n📁 File types found:\n")
	for ext, count := range fileTypes {
		status := "✓ content extracted + AI analysis"
		fmt.Printf("   %s: %d files (%s)\n", ext, count, status)
	}

	fmt.Printf("\n🚀 Starting file processing and AI analysis...\n")
	fmt.Printf("   Note: This will make API calls to analyze each file's content\n")
	fmt.Printf("   API endpoint: %s\n", a.config.APIBaseURL)

	rootNode, err := a.analyzer.PerformFullAnalysis(targetDir, a.config.Mode)
	if err != nil {
		return fmt.Errorf("error performing full analysis: %w", err)
	}

	fmt.Printf("\n\n✅ File processing and AI analysis complete!\n\n")

	jsonOutput, err := json.MarshalIndent(rootNode, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling to json: %w", err)
	}

	outputFile := filepath.Join(a.config.DefaultOutputDir, a.config.JSONOutputFile)
	err = os.WriteFile(outputFile, jsonOutput, 0644)
	if err != nil {
		return fmt.Errorf("error writing json to file: %w", err)
	}

	markdownOutput := analyzer.GenerateMarkdownOutput(rootNode)
	markdownFile := filepath.Join(a.config.DefaultOutputDir, a.config.MarkdownOutputFile)
	err = os.WriteFile(markdownFile, []byte(markdownOutput), 0644)
	if err != nil {
		return fmt.Errorf("error writing markdown to file: %w", err)
	}

	jsonFileInfo, err := os.Stat(outputFile)
	if err != nil {
		return fmt.Errorf("error getting json file info: %w", err)
	}

	markdownFileInfo, err := os.Stat(markdownFile)
	if err != nil {
		return fmt.Errorf("error getting markdown file info: %w", err)
	}

	fmt.Printf("JSON output written to %s (size: %d bytes)\n", outputFile, jsonFileInfo.Size())
	fmt.Printf("Markdown output written to %s (size: %d bytes)\n", markdownFile, markdownFileInfo.Size())

	fmt.Println("\n=== Process Completed Successfully ===")
	fmt.Printf("✓ File tree processed\n")
	fmt.Printf("✓ AI analysis completed for all files\n")
	fmt.Printf("✓ AI analysis completed for all folders\n")
	fmt.Printf("✓ JSON output generated: %s (%d bytes)\n", outputFile, jsonFileInfo.Size())
	fmt.Printf("✓ Markdown output generated: %s (%d bytes)\n", markdownFile, markdownFileInfo.Size())

	fmt.Print("\nPress Enter to exit...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadLine()

	return nil
}

func (a *App) PerformArchitectureAnalysis() error {
	fmt.Println("🏗️  Starting architectural analysis...")

	jsonFile := filepath.Join(a.config.DefaultOutputDir, a.config.JSONOutputFile)
	markdownFile := filepath.Join(a.config.DefaultOutputDir, a.config.MarkdownOutputFile)

	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		return fmt.Errorf("%s not found. Please run the tool without flags first to generate the analysis files", jsonFile)
	}
	if _, err := os.Stat(markdownFile); os.IsNotExist(err) {
		return fmt.Errorf("%s not found. Please run the tool without flags first to generate the analysis files", markdownFile)
	}

	jsonContent, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", jsonFile, err)
	}

	mdContent, err := os.ReadFile(markdownFile)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", markdownFile, err)
	}

	combinedContent := fmt.Sprintf("JSON Structure Data:\n%s\n\nMarkdown Tree Visualization:\n%s", string(jsonContent), string(mdContent))

	fmt.Printf("📊 Total content size: %d characters\n", len(combinedContent))

	var analyses []string
	const maxChunkSize = 50000

	aiClient := analyzer.NewAIClient(a.config)

	if len(combinedContent) > maxChunkSize {
		fmt.Printf("📋 Content exceeds %d characters, splitting into chunks...\n", maxChunkSize)

		chunks := make([]string, 0)
		for i := 0; i < len(combinedContent); i += maxChunkSize {
			end := i + maxChunkSize
			if end > len(combinedContent) {
				end = len(combinedContent)
			}
			chunks = append(chunks, combinedContent[i:end])
		}

		fmt.Printf("🔄 Processing %d chunks...\n", len(chunks))

		archiConcurrency := a.config.Concurrency.ArchiAnalysis
		type chunkJob struct {
			idx   int
			chunk string
		}
		cj := make(chan chunkJob, len(chunks))
		cres := make(chan struct {
			idx int
			txt string
			err error
		}, len(chunks))

		for w := 0; w < archiConcurrency; w++ {
			go func() {
				for job := range cj {
					fmt.Printf("🔍 Analyzing chunk %d/%d...\n", job.idx+1, len(chunks))
					analysis, err := aiClient.AnalyzeArchitecture(job.chunk, fmt.Sprintf("chunk_%d.combined", job.idx+1))
					cres <- struct {
						idx int
						txt string
						err error
					}{idx: job.idx, txt: analysis, err: err}
				}
			}()
		}

		for i, chunk := range chunks {
			cj <- chunkJob{idx: i, chunk: chunk}
		}
		close(cj)

		analyses = make([]string, len(chunks))
		for i := 0; i < len(chunks); i++ {
			r := <-cres
			if r.err != nil {
				return fmt.Errorf("error analyzing chunk %d: %v", r.idx+1, r.err)
			}
			analyses[r.idx] = r.txt
		}

		reduced := analyses
		round := 1
		for len(reduced) > 1 {
			fmt.Printf("🔁 Reduction round %d: %d analyses to reduce...\n", round, len(reduced))

			groups := make([][]string, 0)
			var currentGroup []string
			currentLen := 0

			for _, aText := range reduced {
				aLen := len(aText)
				if len(currentGroup) == 0 {
					currentGroup = []string{aText}
					currentLen = aLen + 2 // separator
					continue
				}

				if currentLen+aLen+2 <= maxChunkSize {
					currentGroup = append(currentGroup, aText)
					currentLen += aLen + 2
				} else {
					groups = append(groups, currentGroup)
					currentGroup = []string{aText}
					currentLen = aLen + 2
				}
			}
			if len(currentGroup) > 0 {
				groups = append(groups, currentGroup)
			}

			fmt.Printf("🔄 Combining %d groups in this round...\n", len(groups))

			concurrency := a.config.Concurrency.ReportChunking
			type result struct {
				idx int
				txt string
				err error
			}

			jobs := make(chan struct {
				idx int
				grp []string
			}, len(groups))
			results := make(chan result, len(groups))

			// worker
			for w := 0; w < concurrency; w++ {
				go func() {
					for job := range jobs {
						fmt.Printf("   ➤ Combining group %d/%d (items: %d)...\n", job.idx+1, len(groups), len(job.grp))
						combined, err := aiClient.CombineArchitecturalAnalyses(job.grp)
						results <- result{idx: job.idx, txt: combined, err: err}
					}
				}()
			}

			// enqueue jobs
			for gi, grp := range groups {
				jobs <- struct {
					idx int
					grp []string
				}{idx: gi, grp: grp}
			}
			close(jobs)

			// collect
			nextRound := make([]string, len(groups))
			for i := 0; i < len(groups); i++ {
				res := <-results
				if res.err != nil {
					return fmt.Errorf("error combining group %d analyses: %v", res.idx+1, res.err)
				}
				nextRound[res.idx] = res.txt
			}

			reduced = nextRound
			fmt.Printf("🔄 Reduction round %d complete: %d analyses remain\n", round, len(reduced))
			round++
		}

		if len(reduced) == 0 {
			return fmt.Errorf("no analyses produced during reduction")
		}

		finalAnalysis := reduced[0]

		err = os.WriteFile(filepath.Join(a.config.DefaultOutputDir, a.config.ReportOutputFile), []byte(finalAnalysis), 0644)
		if err != nil {
			return fmt.Errorf("error writing report: %v", err)
		}

	} else {
		fmt.Println("📋 Content size is manageable, processing as single analysis...")

		analysis, err := aiClient.AnalyzeArchitecture(combinedContent, "output.json and output.md")
		if err != nil {
			return fmt.Errorf("error analyzing architecture: %v", err)
		}

		err = os.WriteFile(filepath.Join(a.config.DefaultOutputDir, a.config.ReportOutputFile), []byte(analysis), 0644)
		if err != nil {
			return fmt.Errorf("error writing report: %v", err)
		}
	}

	fmt.Println("✅ Architectural analysis complete!")
	fmt.Printf("📄 Report saved to: %s\n", filepath.Join(a.config.DefaultOutputDir, a.config.ReportOutputFile))

	return nil
}

func (a *App) isExtractableFile(ext string) bool {
	extractableExts := []string{
		".txt", ".md", ".go", ".js", ".py", ".java", ".c", ".cpp", ".h", ".hpp",
		".css", ".html", ".xml", ".json", ".yaml", ".yml", ".toml", ".ini",
		".cfg", ".conf", ".docx", ".xlsx", ".xls", ".pdf",
	}

	for _, e := range extractableExts {
		if ext == e {
			return true
		}
	}
	return false
}

func (a *App) isImageFile(ext string) bool {
	imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp"}

	for _, e := range imageExts {
		if ext == e {
			return true
		}
	}
	return false
}

func (a *App) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}
