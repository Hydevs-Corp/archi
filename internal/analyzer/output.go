package analyzer

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func GenerateEstimationMarkdown(estimation *CountEstimation) string {
	var md strings.Builder

	md.WriteString("# File and Folder Estimation\n\n")
	md.WriteString(fmt.Sprintf("**Generated on:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	md.WriteString("## Summary\n\n")
	md.WriteString(fmt.Sprintf("- **Total Files:** %d\n", estimation.TotalFiles))
	md.WriteString(fmt.Sprintf("- **Total Folders:** %d\n", estimation.TotalFolders))
	md.WriteString(fmt.Sprintf("- **Estimated Total Execution Time:** %s\n", formatDuration(estimation.TotalEstimatedTime)))
	md.WriteString(fmt.Sprintf("  - File processing time (~4s each): %s\n", formatDuration(time.Duration(estimation.TotalFiles)*4*time.Second)))
	md.WriteString(fmt.Sprintf("  - Folder processing time (~7s each): %s\n", formatDuration(time.Duration(estimation.TotalFolders)*7*time.Second)))
	md.WriteString("\n")

	md.WriteString("## File Types Analysis\n\n")
	md.WriteString("| Extension | Count | Estimated Time |\n")
	md.WriteString("|-----------|-------|----------------|\n")
	for _, stat := range estimation.FileTypeStats {
		md.WriteString(fmt.Sprintf("| %s | %d | %s |\n",
			stat.Extension, stat.Count, formatDuration(stat.EstimatedTime)))
	}
	md.WriteString("\n")

	md.WriteString("## Root Folders Analysis\n\n")
	md.WriteString("| Folder Name | Files | Subfolders | Estimated Time |\n")
	md.WriteString("|-------------|-------|------------|----------------|\n")
	for _, folder := range estimation.RootFolders {
		md.WriteString(fmt.Sprintf("| %s | %d | %d | %s |\n",
			folder.Name, folder.FileCount, folder.SubfolderCount, formatDuration(folder.EstimatedTime)))
	}
	md.WriteString("\n")

	md.WriteString("## Detailed Root Folder Breakdown\n\n")
	for _, folder := range estimation.RootFolders {
		md.WriteString(fmt.Sprintf("### %s\n\n", folder.Name))
		md.WriteString(fmt.Sprintf("- **Path:** `%s`\n", folder.Path))
		md.WriteString(fmt.Sprintf("- **Files:** %d (estimated %s for processing)\n",
			folder.FileCount, formatDuration(time.Duration(folder.FileCount)*4*time.Second)))
		md.WriteString(fmt.Sprintf("- **Subfolders:** %d (estimated %s for processing)\n",
			folder.SubfolderCount, formatDuration(time.Duration(folder.SubfolderCount)*7*time.Second)))
		md.WriteString(fmt.Sprintf("- **Total estimated time for this folder:** %s\n\n", formatDuration(folder.EstimatedTime)))
	}

	md.WriteString("---\n\n")
	md.WriteString("*This estimation calculates execution time based on 4 seconds per file and 7 seconds per folder for AI analysis.*\n")

	return md.String()
}

func GenerateMarkdownOutput(rootNode *Node) string {
	var markdown strings.Builder

	markdown.WriteString("# Directory Tree Analysis\n\n")
	markdown.WriteString("This document shows the analyzed directory structure with AI-generated descriptions.\n\n")
	markdown.WriteString("## Tree Structure\n\n")

	markdown.WriteString("```\n")
	markdown.WriteString(generateMarkdownTree(rootNode, 0, true, []bool{}))
	markdown.WriteString("```\n\n")

	markdown.WriteString("## Legend\n\n")
	markdown.WriteString("- 📁 Directory\n")
	markdown.WriteString("- 📄 Text/Generic file\n")
	markdown.WriteString("- 🖼️ Image file\n")
	markdown.WriteString("- 📋 PDF document\n")
	markdown.WriteString("- 📝 Word document\n")
	markdown.WriteString("- 📊 Excel spreadsheet\n")
	markdown.WriteString("- 🐹 Go source file\n")
	markdown.WriteString("- 📖 Markdown file\n")
	markdown.WriteString("- 📊 JSON file\n\n")

	markdown.WriteString("*Descriptions are AI-generated based on file content analysis.*\n")

	return markdown.String()
}

func generateMarkdownTree(node *Node, depth int, isLast bool, parentPrefixes []bool) string {
	var markdown strings.Builder

	prefix := ""
	for i := 0; i < depth; i++ {
		if i < len(parentPrefixes) && parentPrefixes[i] {
			prefix += "    "
		} else {
			prefix += "│   "
		}
	}

	if depth > 0 {
		if isLast {
			prefix += "└── "
		} else {
			prefix += "├── "
		}
	}

	icon := "📄"
	if node.Type == "directory" {
		icon = "📁"
	} else {
		ext := strings.ToLower(filepath.Ext(node.Name))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp":
			icon = "🖼️"
		case ".pdf":
			icon = "📋"
		case ".docx", ".doc":
			icon = "📝"
		case ".xlsx", ".xls":
			icon = "📊"
		case ".go":
			icon = "🐹"
		case ".md":
			icon = "📖"
		case ".json":
			icon = "📊"
		}
	}

	markdown.WriteString(fmt.Sprintf("%s%s %s", prefix, icon, node.Name))
	markdown.WriteString("\n")

	if len(node.Children) > 0 {
		newParentPrefixes := make([]bool, len(parentPrefixes)+1)
		copy(newParentPrefixes, parentPrefixes)
		newParentPrefixes[depth] = !isLast

		for i, child := range node.Children {
			isChildLast := i == len(node.Children)-1
			markdown.WriteString(generateMarkdownTree(child, depth+1, isChildLast, newParentPrefixes))
		}
	}

	return markdown.String()
}

func formatDuration(d time.Duration) string {
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
