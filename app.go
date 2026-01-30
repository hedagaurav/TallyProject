package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

// App struct
type App struct {
	ctx context.Context
}

// Log struct
type LogEntry struct {
	Timestamp  string   `json:"timestamp"` // Exact time (HH:MM:SS)
	FileName   string   `json:"fileName"`
	Status     string   `json:"status"` // "Success" or "Failed"
	TotalRows  int      `json:"totalRows"`
	SuccessCnt int      `json:"successCnt"`
	FailedCnt  int      `json:"failedCnt"`
	Errors     []string `json:"errors"` // Error details
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

// ---------------------------------------------------------
// TALLY XML TEMPLATE (Ye Tally ka format hai)
// ---------------------------------------------------------
const voucherXMLTemplate = `
<ENVELOPE>
    <HEADER>
        <TALLYREQUEST>Import Data</TALLYREQUEST>
    </HEADER>
    <BODY>
        <IMPORTDATA>
            <REQUESTDESC>
                <REPORTNAME>Vouchers</REPORTNAME>
            </REQUESTDESC>
            <REQUESTDATA>
                <TALLYMESSAGE xmlns:UDF="TallyUDF">
                    <VOUCHER VCHTYPE="Payment" ACTION="Create" OBJVIEW="Accounting Voucher View">
                        <DATE>%s</DATE>
                        <VOUCHERTYPENAME>Payment</VOUCHERTYPENAME>
                        <VOUCHERNUMBER>%s</VOUCHERNUMBER>
                        <NARRATION>%s</NARRATION>
                        
                        <LEDGERENTRIES.LIST>
                            <LEDGERNAME>%s</LEDGERNAME>
                            <ISDEEMEDPOSITIVE>Yes</ISDEEMEDPOSITIVE>
                            <AMOUNT>-%s</AMOUNT>
                        </LEDGERENTRIES.LIST>
                        
                        <LEDGERENTRIES.LIST>
                            <LEDGERNAME>%s</LEDGERNAME>
                            <ISDEEMEDPOSITIVE>No</ISDEEMEDPOSITIVE>
                            <AMOUNT>%s</AMOUNT>
                        </LEDGERENTRIES.LIST>
                    </VOUCHER>
                </TALLYMESSAGE>
            </REQUESTDATA>
        </IMPORTDATA>
    </BODY>
</ENVELOPE>`

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

func (a *App) UploadToTally(filePath string) string {
	if filePath == "" {
		return "Please select a file first!"
	}

	// Step A: Excel File Open karo
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return "Error: Could not open file. " + err.Error()
	}
	defer f.Close()

	// Step B: Rows Read karo (Sheet1 se)
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return "Error: Could not read rows. " + err.Error()
	}

	successCount := 0
	failCount := 0
	tallyURL := "http://localhost:9000" // Tally Server Address

	errorList := []string{}

	// Step C: Loop chalao (Row by Row)
	for i, row := range rows {
		if i == 0 {
			continue // Header row skip karo
		}

		// Check karo ki row khali to nahi hai (kam se kam 5 column hone chahiye)
		if len(row) < 5 {
			continue
		}

		// Excel se Data nikalo (Columns: Date, VchNo, DebitLedger, CreditLedger, Amount, Narration)
		dateRaw := row[0]
		vchNo := row[1]
		drLedger := row[2]
		crLedger := row[3]
		amount := row[4]
		narration := ""
		if len(row) > 5 {
			narration = row[5]
		}

		// Date format fix karo (2026-01-29 -> 20260129)
		tallyDate := strings.ReplaceAll(dateRaw, "-", "")

		// XML Data taiyar karo
		xmlPayload := fmt.Sprintf(voucherXMLTemplate,
			tallyDate, vchNo, narration, drLedger, amount, crLedger, amount)

		// Tally ko bhejo
		err := sendToTally(tallyURL, xmlPayload)
		if err != nil {
			failCount++
			fmt.Println("Error on Row", i+1, err)
		} else {
			successCount++
		}
	}

	a.SaveLogToDailyFile(filePath, len(rows)-1, successCount, failCount, errorList)

	return fmt.Sprintf("Completed! Success: %d, Failed: %d", successCount, failCount)
	// return fmt.Sprintf("Completed! Rows: %d", len(rows))

}

// Helper Function: Jo XML ko Tally tak lekar jayega
func sendToTally(url, xmlData string) error {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(xmlData)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
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

	newEntry := LogEntry{
		Timestamp:  time.Now().Format(time.TimeOnly),
		FileName:   fileName,
		Status:     status,
		TotalRows:  total,
		SuccessCnt: success,
		FailedCnt:  failed,
		Errors:     errorList,
	}

	// 4. Purana data read karo (Agar file exist karti hai)
	var logs []LogEntry
	fileData, err := os.ReadFile(filePath)
	if err == nil {
		// File hai, to data parse karo
		json.Unmarshal(fileData, &logs)
	}

	// 5. Naya log list me jodo (Naya sabse upar rakhna hai to 'append' ka order badal dein)
	// Abhi hum neeche add kar rahe hain
	// logs = append(logs, newEntry)

	// logs ko prepend karne ke liye
	logs = append([]LogEntry{newEntry}, logs...)

	// 6. File wapas save karo (Indent ke sath taaki padhne me aasaan ho)
	updatedData, _ := json.MarshalIndent(logs, "", "  ")
	os.WriteFile(filePath, updatedData, 0644)
}
