package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/xuri/excelize/v2"
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
	expectedKey := generateHash(id + constant.AppSecret)

	return inputKey == expectedKey
}

// 3. License Activate karne ka function (Jab user key daalega)
func (a *App) ActivateLicense(key string) string {
	id, _ := machineid.ProtectedID("TallyApp")
	expectedKey := generateHash(id + constant.AppSecret)

	if key == expectedKey {
		// Sahi key hai -> File save karo
		os.WriteFile("license.key", []byte(key), 0644)
		return "Success"
	}
	return "Invalid Key! Please contact Admin."
}

// Helper: Hash Generator (SHA256)
func generateHash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
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

func (a *App) ImportPurchaseVoucher(filePath string) string {
	// 1. Log: File Path Check
	fmt.Println("Step 1: File Path mile ->", filePath)

	if filePath == "" {
		return "Please select a file first!"
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return "Error: Could not open Excel file"
	}
	defer f.Close()

	// 2. Log: Sheet Names Check
	sheetList := f.GetSheetList()
	fmt.Println("Step 2: Available Sheets ->", sheetList)

	// Note: Image me sheet ka naam "April-25" dikh raha hai.
	// Agar fix naam hai to "April-25" use karein, ya first sheet utha lein
	sheetName := "April-25"

	// Agar sheet naam dynamic rakhna ho (jo pehli sheet ho wahi utha lo):
	if len(sheetList) > 0 {
		sheetName = sheetList[0]
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Printf("Error: '%s' sheet nahi mili. Available: %v\n", sheetName, sheetList)
		return "Error: Sheet name not found"
	}

	// 3. Log: Total Rows
	fmt.Printf("Step 3: Total Rows Found -> %d\n", len(rows))

	success := 0
	skipped := 0
	errors := 0

	// Loop Starts
	for i, row := range rows {
		// Headers Skip (Row 1-2 headers nahi hain, data Row 63 se hai, lekin
		// "D.V" filter headers ko apne aap hata dega, so bas safe indexing chahiye)
		if i < 2 {
			continue
		}

		// 4. Log: Row Length Check (UPDATED for Column V)
		// Column V ka index 21 hai, isliye length kam se kam 22 honi chahiye
		if len(row) < 22 {
			// Sirf tab log print karo agar ye row "D.V" wali ho sakti thi
			if len(row) > 0 && row[0] == "D.V" {
				fmt.Printf("⚠️ Row %d SKIPPED due to length < 22 (Length: %d)\n", i+1, len(row))
			}
			continue
		}

		// 5. Log: Company Filter
		companyCode := row[0]
		if companyCode != "D.V" {
			// Har row ka log print mat karo warna console bhar jayega, sirf error debugging ke liye rakho
			skipped++
			continue
		}

		// --- Data Parsing ---
		fmt.Printf("Processing Row %d for D.V...\n", i+1)

		fmt.Printf("Excel date: %s", row[1])

		// Date Parsing (DD-MM-YYYY -> YYYYMMDD)
		excelFormattedDate, err := helper.ParseDateSmart(row[1])
		if err != nil {
			fmt.Printf("❌ Row %d Date Error: '%s' samajh nahi aayi -> %v\n", i+1, row[1], err)
			errors++
			continue
		}

		fmt.Printf("excelFormattedDate: %s", excelFormattedDate)

		// Safe Data Extraction (Indices Updated based on Image)
		entry := model.PurchaseEntry{
			PartyName: row[2],                    // Col C (Party Name)
			InvoiceNo: row[4],                    // Col E (INVOICE)
			ItemName:  row[5],                    // Col F (Brand)
			Qty:       helper.ParseFloat(row[7]), // Col H (QTY)
			Rate:      helper.ParseFloat(row[8]), // Col I (RATE)

			// Tax Columns Updated (M, N, O)
			IGSTAmount: helper.ParseFloat(row[12]), // Col M (Index 12)
			CGSTAmount: helper.ParseFloat(row[13]), // Col N (Index 13)
			SGSTAmount: helper.ParseFloat(row[14]), // Col O (Index 14)

			// Total Amount Updated (Column V)
			TotalBill: helper.ParseFloat(row[21]), // Col V (Index 21)

			Date: excelFormattedDate,
		}

		entry.Amount = entry.Qty * entry.Rate
		entry.GUID = helper.GenerateGUID(entry.PartyName, entry.InvoiceNo, entry.Date)

		// --- XML Logic ---
		var taxXML string
		if entry.IGSTAmount > 0 {
			taxXML = fmt.Sprintf(constant.PurchaseIGSTTemplate, entry.IGSTAmount)
		} else {
			taxXML = fmt.Sprintf(constant.PurchaseCGSTSGSTTemplate, entry.CGSTAmount, entry.SGSTAmount)
		}

		finalXML := fmt.Sprintf(constant.PurchaseXMLTemplate,
			entry.Date,                        // DATE
			entry.Date,                        // REFERENCEDATE
			entry.GUID,                        // GUID
			helper.EscapeXML(entry.InvoiceNo), // VOUCHERNUMBER
			helper.EscapeXML(entry.InvoiceNo), // REFERENCE
			helper.EscapeXML(entry.PartyName), // PARTYLEDGERNAME

			// --- Inventory Block ---
			helper.EscapeXML(entry.ItemName), // STOCKITEMNAME
			entry.Rate,                       // RATE
			entry.Qty,                        // ACTUALQTY
			entry.Qty,                        // BILLEDQTY
			entry.Amount,                     // AMOUNT

			// --- Batch Allocation Block (REPEAT VALUES) ---
			entry.Amount, // Batch AMOUNT
			entry.Qty,    // Batch ACTUALQTY
			entry.Qty,    // Batch BILLEDQTY

			// --- Accounting Allocation Block (REPEAT VALUES) ---
			entry.Amount, // Accounting AMOUNT

			// --- Party Ledger ---
			helper.EscapeXML(entry.PartyName), // LEDGERNAME
			entry.TotalBill,                   // AMOUNT (Total Positive)

			// --- Tax XML ---
			taxXML)

		fmt.Println("SENDING XML:", finalXML)

		// 6. Log: Sending to Tally
		tallyURL := "http://localhost:9000"
		resp, err := helper.SendToTally(tallyURL, finalXML)

		if err != nil {
			fmt.Printf("❌ Row %d Network Error: %v\n", i+1, err)
			errors++
		} else if strings.Contains(resp, "<CREATED>1</CREATED>") {
			fmt.Printf("✅ Row %d Success!\n", i+1)
			success++
		} else {
			// Duplicate GUID error ya koi aur logic error check karne ke liye:
			fmt.Printf("❌ Row %d Tally Rejected: %s\n", i+1, resp)
			errors++
		}
	}

	return fmt.Sprintf("Import Complete! Success: %d, Skipped: %d, Failed: %d", success, skipped, errors)
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
