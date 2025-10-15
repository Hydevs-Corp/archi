package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"archi/internal/config"
)

type Analyzer struct {
	config   *config.Config
	aiClient *AIClient
}

func New(cfg *config.Config) *Analyzer {
	return &Analyzer{
		config:   cfg,
		aiClient: NewAIClient(cfg),
	}
}

func (a *Analyzer) PerformCountAnalysis(rootPath string) (*CountEstimation, error) {
	estimation := &CountEstimation{
		FileTypeStats: make([]FileTypeStats, 0),
		RootFolders:   make([]FolderStats, 0),
	}

	fileTypeCounts := make(map[string]int)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			estimation.TotalFolders++
		} else {
			estimation.TotalFiles++
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext == "" {
				ext = "no extension"
			}
			fileTypeCounts[ext]++
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	for ext, count := range fileTypeCounts {
		estimatedTime := time.Duration(count) * 4 * time.Second
		estimation.FileTypeStats = append(estimation.FileTypeStats, FileTypeStats{
			Extension:     ext,
			Count:         count,
			EstimatedTime: estimatedTime,
		})
	}

	rootDir, err := os.Open(rootPath)
	if err != nil {
		return nil, err
	}
	defer rootDir.Close()

	entries, err := rootDir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			folderPath := filepath.Join(rootPath, entry.Name())
			stats, err := a.calculateFolderStats(folderPath, entry.Name())
			if err != nil {
				continue
			}
			estimation.RootFolders = append(estimation.RootFolders, *stats)
		}
	}

	fileTime := time.Duration(estimation.TotalFiles) * 4 * time.Second
	folderTime := time.Duration(estimation.TotalFolders) * 7 * time.Second
	estimation.TotalEstimatedTime = fileTime + folderTime

	return estimation, nil
}

func (a *Analyzer) calculateFolderStats(folderPath, folderName string) (*FolderStats, error) {
	stats := &FolderStats{
		Name: folderName,
		Path: folderPath,
	}

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != folderPath {
			stats.SubfolderCount++
		} else if !info.IsDir() {
			stats.FileCount++
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	fileTime := time.Duration(stats.FileCount) * 4 * time.Second
	folderTime := time.Duration(stats.SubfolderCount+1) * 7 * time.Second // +1 for the folder itself
	stats.EstimatedTime = fileTime + folderTime

	return stats, nil
}

func (a *Analyzer) PerformFullAnalysis(rootPath string, onlyFolders, noContent bool) (*Node, error) {
	// Normalize the root path to avoid trailing-slash mismatches when linking parent/child nodes
	rootPath = filepath.Clean(rootPath)
	nodes := make(map[string]*Node)
	var rootNode *Node
	currentFile := 0

	var totalFiles int
	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && (!onlyFolders) {
			totalFiles++
		}
		return nil
	})

	var fileNodes []*Node
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if onlyFolders && !info.IsDir() {
			return nil
		}

		// Normalize the current path for consistent map keys and comparisons
		cleanPath := filepath.Clean(path)

		node := &Node{
			Path: cleanPath,
			Name: info.Name(),
		}

		if info.IsDir() {
			node.Type = "directory"
		} else {
			node.Type = "file"
		}

		if node.Type == "file" {
			fileNodes = append(fileNodes, node)
		}

		nodes[cleanPath] = node

		if cleanPath == rootPath {
			rootNode = node
		} else {
			parentPath := filepath.Clean(filepath.Dir(cleanPath))
			if parent, exists := nodes[parentPath]; exists {
				parent.Children = append(parent.Children, node)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if !onlyFolders {
		fmt.Printf("\n\n📦 Analyzing files in batches of %d...\n", a.config.BatchSize)
		total := len(fileNodes)
		currentFile = 0
		for i := 0; i < total; i += a.config.BatchSize {
			end := i + a.config.BatchSize
			if end > total {
				end = total
			}

			batch := fileNodes[i:end]
			var wg sync.WaitGroup
			wg.Add(len(batch))

			for _, n := range batch {
				n := n
				path := n.Path
				info, err := os.Stat(path)
				if err != nil {
					fmt.Printf("\n⚠️  Skipping file %s: %v\n", path, err)
					wg.Done()
					continue
				}

				go func() {
					defer wg.Done()
					ext := strings.ToLower(filepath.Ext(info.Name()))
					switch ext {
					case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp":
						desc, err := a.aiClient.AnalyzeImage(path)
						if err != nil {
							fmt.Printf("\n⚠️  Error analyzing image %s: %v\n", path, err)
							return
						}
						n.Description = fmt.Sprintf("Image analysis: %s", desc)
					default:
						content, err := a.extractFileContent(path, info)
						if err != nil || content == "" {
							return
						}
						if len(content) > 5000 {
							content = content[:5000]
						}
						if !noContent {
							n.Content = content
						}
						desc, err := a.aiClient.AnalyzeFileContent(content, info.Name())
						if err != nil {
							fmt.Printf("\n⚠️  Error analyzing file %s: %v\n", path, err)
							return
						}
						n.Description = desc
					}
				}()
			}
			wg.Wait()
			currentFile += len(batch)
			a.printProgressBar(currentFile, total, "📄 Processing files:")
			time.Sleep(a.config.RequestDelay)
		}
	}

	fmt.Printf("\n\n🗂️  Starting folder description generation...\n")
	totalFolders := a.countFolders(rootNode)
	fmt.Printf("   Found %d folders to analyze\n", totalFolders)

	var folderNodes []*Node
	var collect func(n *Node)
	collect = func(n *Node) {
		if n.Type == "directory" {
			folderNodes = append(folderNodes, n)
		}
		for _, ch := range n.Children {
			collect(ch)
		}
	}
	collect(rootNode)

	currentFolder := 0
	for i := 0; i < len(folderNodes); i += a.config.BatchSize {
		end := i + a.config.BatchSize
		if end > len(folderNodes) {
			end = len(folderNodes)
		}
		batch := folderNodes[i:end]
		var wg sync.WaitGroup
		wg.Add(len(batch))
		for _, n := range batch {
			n := n
			go func() {
				defer wg.Done()
				desc, err := a.aiClient.AnalyzeFolderContent(n)
				if err != nil {
					fmt.Printf("\n⚠️  Error analyzing folder %s: %v\n", n.Path, err)
					return
				}
				n.Description = desc
			}()
		}
		wg.Wait()
		currentFolder += len(batch)
		a.printProgressBar(currentFolder, totalFolders, "📁 Processing folders:")
		time.Sleep(a.config.RequestDelay)
	}

	return rootNode, nil
}

func (a *Analyzer) extractFileContent(path string, info os.FileInfo) (string, error) {
	ext := strings.ToLower(filepath.Ext(info.Name()))

	isImageFile := ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".bmp" || ext == ".webp"

	if !isImageFile && info.Size() > a.config.MaxFileSize {
		return "", fmt.Errorf("file too large")
	}

	switch ext {
	case ".txt", ".md", ".go", ".js", ".py", ".java", ".c", ".cpp", ".h", ".hpp", ".css", ".html", ".xml", ".json", ".yaml", ".yml", ".toml", ".ini", ".cfg", ".conf":
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(content), nil

	case ".docx":
		return ReadDocx(path)

	case ".xlsx", ".xls":
		return ReadXlsx(path)

	case ".pdf":
		return ReadPdf(path)

	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp":
		// We don't analyze images here to keep text-only content extraction.
		// Image analysis is triggered via AI client when appropriate elsewhere.
		return "", nil

	default:
		return "", fmt.Errorf("unsupported file type")
	}
}

func (a *Analyzer) processFoldersForDescription(node *Node, currentFolder *int, totalFolders int) error {
	if node.Type == "directory" {
		*currentFolder++
		a.printProgressBar(*currentFolder, totalFolders, "📁 Processing folders:")

		description, err := a.aiClient.AnalyzeFolderContent(node)
		if err != nil {
			fmt.Printf("\n⚠️  Error analyzing folder %s: %v\n", node.Path, err)
		} else {
			node.Description = description
		}
		a.printProgressBar(*currentFolder, totalFolders, "📁 Processing folders:")

		time.Sleep(a.config.RequestDelay)
	}

	for _, child := range node.Children {
		if err := a.processFoldersForDescription(child, currentFolder, totalFolders); err != nil {
			return err
		}
	}

	return nil
}

func (a *Analyzer) countFolders(node *Node) int {
	if node == nil {
		return 0
	}
	count := 0
	if node.Type == "directory" {
		count = 1
	}
	for _, child := range node.Children {
		count += a.countFolders(child)
	}
	return count
}

func (a *Analyzer) printProgressBar(current, total int, prefix string) {
	if total == 0 {
		return
	}

	barWidth := 40
	percentage := float64(current) / float64(total) * 100
	filled := int(float64(barWidth) * float64(current) / float64(total))

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"

	fmt.Printf("\r%s %s %.1f%% (%d/%d)", prefix, bar, percentage, current, total)
}
