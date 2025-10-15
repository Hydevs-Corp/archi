package analyzer

import (
	"time"
)

type Node struct {
	Path        string  `json:"path"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Content     string  `json:"content,omitempty"`
	Description string  `json:"description,omitempty"`
	Children    []*Node `json:"children,omitempty"`
}

type FileTypeStats struct {
	Extension     string        `json:"extension"`
	Count         int           `json:"count"`
	EstimatedTime time.Duration `json:"estimatedTime"`
}

type FolderStats struct {
	Name           string        `json:"name"`
	Path           string        `json:"path"`
	FileCount      int           `json:"fileCount"`
	SubfolderCount int           `json:"subfolderCount"`
	EstimatedTime  time.Duration `json:"estimatedTime"`
}

type CountEstimation struct {
	TotalFiles         int             `json:"totalFiles"`
	TotalFolders       int             `json:"totalFolders"`
	FileTypeStats      []FileTypeStats `json:"fileTypeStats"`
	RootFolders        []FolderStats   `json:"rootFolders"`
	TotalEstimatedTime time.Duration   `json:"totalEstimatedTime"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	History []ChatMessage `json:"history"`
	// Model can be either a string (for Mistral-only shortcut) or an array of provider/model objects
	Model   interface{}   `json:"model"`
}

type ChatResponse struct {
	Response    string `json:"response"`
	QueueLength int    `json:"queueLength"`
}

type ImageRequest struct {
	Image string `json:"image"`
	// Model can be either a string (for Mistral-only shortcut) or an array of provider/model objects
	Model interface{} `json:"model"`
}

type ImageResponse struct {
	Analysis    string `json:"analysis"`
	QueueLength int    `json:"queueLength"`
}
