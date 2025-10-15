package analyzer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"

	"archi/internal/config"

	"github.com/nfnt/resize"
)

type AIClient struct {
	config *config.Config
}

func NewAIClient(cfg *config.Config) *AIClient {
	return &AIClient{config: cfg}
}

func (c *AIClient) compressImage(imagePath string) ([]byte, error) {
	const maxSizeBytes = 5 * 1024 * 1024 // 5MB limit

	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error opening image file: %v", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %v", err)
	}

	bounds := img.Bounds()
	originalWidth := bounds.Max.X - bounds.Min.X
	originalHeight := bounds.Max.Y - bounds.Min.Y

	currentWidth := uint(originalWidth)
	currentHeight := uint(originalHeight)
	quality := 85

	for {
		var resizedImg image.Image
		if currentWidth == uint(originalWidth) && currentHeight == uint(originalHeight) {
			resizedImg = img
		} else {
			resizedImg = resize.Resize(currentWidth, currentHeight, img, resize.Lanczos3)
		}

		var buf bytes.Buffer
		var encodeErr error

		switch strings.ToLower(format) {
		case "jpeg", "jpg":
			encodeErr = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: quality})
		case "png":
			encodeErr = png.Encode(&buf, resizedImg)
		default:
			encodeErr = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: quality})
		}

		if encodeErr != nil {
			return nil, fmt.Errorf("error encoding image: %v", encodeErr)
		}

		if buf.Len() <= maxSizeBytes {
			return buf.Bytes(), nil
		}

		if quality > 50 {
			quality -= 10
		} else {
			currentWidth = uint(float64(currentWidth) * 0.9)
			currentHeight = uint(float64(currentHeight) * 0.9)
			quality = 85
		}

		if currentWidth < 100 || currentHeight < 100 {
			return buf.Bytes(), nil
		}
	}
}

func (c *AIClient) AnalyzeFileContent(content, filename string) (string, error) {
	var model interface{}
	if len(c.config.FileAnalysisModels) > 0 {
		model = c.config.FileAnalysisModels
	} else {
		model = c.config.FileAnalysisModel
	}
	request := ChatRequest{
		History: []ChatMessage{
			{
				Role:    "user",
				Content: fmt.Sprintf("Please describe the content of this file named '%s' in 250 words maximum based on the following content (first 5000 characters):\n\n%s", filename, content),
			},
		},
		Model: model,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := http.Post(c.config.APIBaseURL+"/ask", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return response.Response, nil
}

func (c *AIClient) AnalyzeImage(imagePath string) (string, error) {
	imageData, err := c.compressImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("error compressing image: %v", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	var model interface{}
	if len(c.config.ImageAnalysisModels) > 0 {
		model = c.config.ImageAnalysisModels
	} else {
		model = c.config.ImageAnalysisModel
	}
	request := ImageRequest{Image: base64Image, Model: model}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := http.Post(c.config.APIBaseURL+"/analyze-image", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return response.Analysis, nil
}

func (c *AIClient) AnalyzeFolderContent(node *Node) (string, error) {
	if node.Type != "directory" || len(node.Children) == 0 {
		return "", nil
	}

	childrenToAnalyze := node.Children
	if len(childrenToAnalyze) > 20 {
		childrenToAnalyze = childrenToAnalyze[:20]
	}

	contentBuilder := fmt.Sprintf("Folder: %s\nContents:\n", node.Name)

	for _, child := range childrenToAnalyze {
		if child.Type == "directory" {
			contentBuilder += fmt.Sprintf("📁 %s/ (directory", child.Name)
			if len(child.Children) > 0 {
				contentBuilder += fmt.Sprintf(" with %d items", len(child.Children))
			}
			contentBuilder += ")\n"
		} else {
			contentBuilder += fmt.Sprintf("📄 %s", child.Name)
			if child.Description != "" {
				desc := child.Description
				if len(desc) > 100 {
					desc = desc[:100] + "..."
				}
				contentBuilder += fmt.Sprintf(" - %s", desc)
			}
			contentBuilder += "\n"
		}
	}

	if len(node.Children) > 20 {
		contentBuilder += fmt.Sprintf("... and %d more items\n", len(node.Children)-20)
	}

	prompt := fmt.Sprintf("Please describe this folder named '%s' in 250 words maximum based on its contents below:\n\n%s", node.Name, contentBuilder)

	var model interface{}
	if len(c.config.FolderAnalysisModels) > 0 {
		model = c.config.FolderAnalysisModels
	} else {
		model = c.config.FolderAnalysisModel
	}
	request := ChatRequest{
		History: []ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Model: model,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := http.Post(c.config.APIBaseURL+"/ask", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return response.Response, nil
}

func (c *AIClient) AnalyzeArchitecture(content, filename string) (string, error) {
	var model interface{}
	if len(c.config.ArchitectureAnalysisModels) > 0 {
		model = c.config.ArchitectureAnalysisModels
	} else {
		model = c.config.ArchitectureAnalysisModel
	}
	request := ChatRequest{
		History: []ChatMessage{
			{
				Role:    "user",
				Content: fmt.Sprintf("Please analyze the software architecture of this project based on the provided file structure and descriptions from '%s'. Provide detailed recommendations for better architecture, including:\n\n1. Current architecture analysis\n2. Identified issues and anti-patterns\n3. Suggested improvements\n4. Recommended folder structure\n5. Best practices recommendations\n6. Technology stack optimization suggestions\n\nContent to analyze:\n%s", filename, content),
			},
		},
		Model: model,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := http.Post(c.config.APIBaseURL+"/ask", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return response.Response, nil
}

func (c *AIClient) CombineArchitecturalAnalyses(analyses []string) (string, error) {
	analysesText := ""
	for _, analysis := range analyses {
		analysesText += analysis + "\n\n---\n\n"
	}

	var model interface{}
	if len(c.config.ArchitectureAnalysisModels) > 0 {
		model = c.config.ArchitectureAnalysisModels
	} else {
		model = c.config.ArchitectureAnalysisModel
	}
	request := ChatRequest{
		History: []ChatMessage{
			{
				Role:    "user",
				Content: fmt.Sprintf("Please combine and synthesize the following architectural analyses into a comprehensive final report. Create a cohesive architectural recommendation document that:\n\n1. Consolidates all findings into a unified analysis\n2. Removes redundancy while preserving important details\n3. Provides a clear executive summary\n4. Presents actionable recommendations in priority order\n5. Includes a proposed implementation roadmap\n\nAnalyses to combine:\n\n%s", analysesText),
			},
		},
		Model: model,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := http.Post(c.config.APIBaseURL+"/ask", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return response.Response, nil
}
