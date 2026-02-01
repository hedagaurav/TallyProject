package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	constant "TallyProject/constants"
	helper "TallyProject/helpers"
	model "TallyProject/models"

	"github.com/denisbrodbeck/machineid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// 1. Machine ID lene ka function (Frontend ko dikhane ke liye)
func (a *App) GetMachineID() string {
	id, err := machineid.ProtectedID("TallyApp")
	if err != nil {
		return "UnknownID"
	}
	return id
}

// 2. License Check karne ka function (Startup par call hoga)
func (a *App) CheckLicense() bool {
	// A. Machine ID nikalo
	id, err := machineid.ProtectedID("TallyApp")
	if err != nil {
		return false
	}

	// B. License file padho
	keyData, err := os.ReadFile("license.key")
	if err != nil {
		return false // File nahi mili matlab license nahi hai
	}
	inputKey := strings.TrimSpace(string(keyData))

	// C. Valid Key generate karke match karo
	expectedKey := helper.GenerateHash(id + constant.AppSecret)

	return inputKey == expectedKey
}

// 3. License Activate karne ka function (Jab user key daalega)
func (a *App) ActivateLicense(key string) string {
	id, _ := machineid.ProtectedID("TallyApp")
	expectedKey := helper.GenerateHash(id + constant.AppSecret)

	if key == expectedKey {
		// Sahi key hai -> File save karo
		os.WriteFile("license.key", []byte(key), 0644)
		return "Success"
	}
	return "Invalid Key! Please contact Admin."
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// 1. File Browse Function (Frontend se call hoga)
func (a *App) SelectExcelFile() string {
	// Native OS File Dialog open karega
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Excel File for Tally",
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel Files", Pattern: "*.xlsx"},
		},
	})

	if err != nil {
		return ""
	}
	return selection // Selected file ka path wapas bhejega
}

func (a *App) SaveLogToDailyFile(fileName string, total, success, failed int, errorList []string) {
	// 1. Logs folder ka path set karo
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.Mkdir(logDir, 0755) // Agar folder nahi hai to banao
	}

	// 2. Aaj ki date ke hisaab se filename (e.g., 2026-01-30.json)
	today := time.Now().Format(time.DateOnly)
	filePath := filepath.Join(logDir, today+".json")

	// 3. Naya Data taiyar karo
	status := "Success"
	if failed > 0 {
		status = "Partial Failure"
	}
	if success == 0 && failed > 0 {
		status = "Failed"
	}

	newEntry := model.LogEntry{
		Timestamp:  time.Now().Format(time.TimeOnly),
		FileName:   fileName,
		Status:     status,
		TotalRows:  total,
		SuccessCnt: success,
		FailedCnt:  failed,
		Errors:     errorList,
	}

	// 4. Purana data read karo (Agar file exist karti hai)
	var logs []model.LogEntry
	fileData, err := os.ReadFile(filePath)
	if err == nil {
		// File hai, to data parse karo
		json.Unmarshal(fileData, &logs)
	}

	// 5. Naya log list me jodo (Naya sabse upar rakhna hai to 'append' ka order badal dein)
	// Abhi hum neeche add kar rahe hain
	// logs = append(logs, newEntry)

	// logs ko prepend karne ke liye
	logs = append([]model.LogEntry{newEntry}, logs...)

	// 6. File wapas save karo (Indent ke sath taaki padhne me aasaan ho)
	updatedData, _ := json.MarshalIndent(logs, "", "  ")
	os.WriteFile(filePath, updatedData, 0644)
}
