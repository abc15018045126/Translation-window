package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed llama-bin2/*
var llamaBin embed.FS

// App struct
type App struct {
	ctx       context.Context
	llamaCmd  *exec.Cmd
	llamaPort string
}

// NewApp creates a new App application struct
func NewApp() *App {
	a := &App{}
	return a
}

func extractFile(fs embed.FS, src string, dest string) {
	data, err := fs.ReadFile(src)
	if err != nil {
		fmt.Println("Error reading embedded file:", err)
		return
	}
	if st, err := os.Stat(dest); os.IsNotExist(err) || st.Size() != int64(len(data)) {
		os.WriteFile(dest, data, 0755)
	}
}

func (a *App) initAI(customModelPath string) {
	tempDir := filepath.Join(os.TempDir(), "translations_llama_v2")
	os.RemoveAll(filepath.Join(os.TempDir(), "translations_llama"))
	os.RemoveAll(tempDir)
	os.MkdirAll(tempDir, 0755)

	entries, err := llamaBin.ReadDir("llama-bin2")
	if err != nil {
		fmt.Println("Error reading llama-bin2 dir:", err)
	}
	fmt.Printf("Found %d entries in llama-bin2\n", len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			src := "llama-bin2/" + entry.Name()
			dest := filepath.Join(tempDir, entry.Name())
			extractFile(llamaBin, src, dest)
		}
	}

	llamaPath := filepath.Join(tempDir, "llama-server.exe")


	modelPath := customModelPath
	if modelPath == "" {
		models := a.GetModels()
		if len(models) > 0 {
			modelPath = filepath.Join("C:\\Users\\abc15\\Documents\\GitHub\\Translation-window\\modal", models[0])
		} else {
			fmt.Println("No models found in modal folder")
			return
		}
	}

	a.llamaPort = "49213"
	
	if a.llamaCmd != nil && a.llamaCmd.Process != nil {
		a.llamaCmd.Process.Kill()
		a.llamaCmd.Wait()
	} else {
		exec.Command("taskkill", "/F", "/IM", "llama-server.exe").Run()
	}
	
	time.Sleep(1 * time.Second)

	a.llamaCmd = exec.Command(llamaPath, "-m", modelPath, "--port", a.llamaPort)
	a.llamaCmd.Dir = tempDir
	err = a.llamaCmd.Start()
	if err != nil {
		fmt.Println("Failed to start AI server:", err)
		return
	}

	// wait for it to be ready
	for i := 0; i < 40; i++ {
		time.Sleep(500 * time.Millisecond)
		resp, err := http.Get("http://127.0.0.1:" + a.llamaPort + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 || resp.StatusCode == 503 {
				fmt.Println("AI server is ready!")
				break
			}
		}
	}
}

// GetModels returns a list of gguf files in the modal directory
func (a *App) GetModels() []string {
	dir := filepath.Join("C:\\Users\\abc15\\Documents\\GitHub\\Translation-window\\modal")
	files, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}

	var models []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".gguf") {
			models = append(models, f.Name())
		}
	}
	return models
}

// SwitchModel switches the AI to the specified model file name from the modal folder
func (a *App) SwitchModel(modelName string) string {
	if modelName == "" {
		return ""
	}

	modelPath := filepath.Join("C:\\Users\\abc15\\Documents\\GitHub\\Translation-window\\modal", modelName)
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return ""
	}
	
	go a.initAI(modelPath)
	return modelName
}

func (a *App) shutdown(ctx context.Context) {
	if a.llamaCmd != nil && a.llamaCmd.Process != nil {
		a.llamaCmd.Process.Kill()
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.initAI("")
}

func (a *App) fetchAITranslation(query string) string {
	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type Payload struct {
		Messages    []Message `json:"messages"`
		Temperature float64   `json:"temperature"`
		MaxTokens   int       `json:"max_tokens"`
	}

	isChinese := false
	for _, r := range query {
		if r > 127 {
			isChinese = true
			break
		}
	}

	systemPrompt := "You are a highly accurate, precise translator. "
	if isChinese {
		systemPrompt += "Translate the following Chinese text into fluent and idiomatic English. Respond with ONLY the final English translation, without any explanations, prefixes, or quotes."
	} else {
		systemPrompt += "Translate the following English text into fluent and idiomatic Simplified Chinese (简体中文). Respond with ONLY the final Chinese translation, without any explanations, prefixes, or quotes."
	}

	payload := Payload{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: query},
		},
		Temperature: 0.1,
		MaxTokens:   512,
	}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "http://127.0.0.1:"+a.llamaPort+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Allow enough time for local AI translation
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "⚠️ AI Service Error or Loading Model..."
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var respData struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(bodyBytes, &respData); err == nil && len(respData.Choices) > 0 {
		return strings.TrimSpace(respData.Choices[0].Message.Content)
	}

	return "⚠️ AI failed to generate translation."
}

// Translate uses the AI model to translate the text directly
func (a *App) Translate(query string) string {
	query = strings.TrimSpace(query)
	if len(query) == 0 {
		return ""
	}

	return a.fetchAITranslation(query)
}
