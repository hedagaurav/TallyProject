package main

import (
	"TallyProject/constants"
	"TallyProject/helpers"
	"TallyProject/models"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

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
		excelFormattedDate, err := helpers.ParseDateSmart(row[1])
		if err != nil {
			fmt.Printf("❌ Row %d Date Error: '%s' samajh nahi aayi -> %v\n", i+1, row[1], err)
			errors++
			continue
		}

		fmt.Printf("excelFormattedDate: %s", excelFormattedDate)

		// Safe Data Extraction (Indices Updated based on Image)
		entry := models.PurchaseEntry{
			Date:      excelFormattedDate,
			PartyName: row[2],                     // Col C (Party Name)
			InvoiceNo: row[4],                     // Col E (INVOICE)
			ItemName:  row[5],                     // Col F (Brand)
			Qty:       helpers.ParseFloat(row[7]), // Col H (QTY)
			Rate:      helpers.ParseFloat(row[8]), // Col I (RATE)

			// Tax Columns Updated (M, N, O)
			IGSTAmount: helpers.ParseFloat(row[12]), // Col M (Index 12)
			CGSTAmount: helpers.ParseFloat(row[13]), // Col N (Index 13)
			SGSTAmount: helpers.ParseFloat(row[14]), // Col O (Index 14)

			// Total Amount Updated (Column V)
			TotalBill: helpers.ParseFloat(row[21]), // Col V (Index 21)

		}

		entry.Amount = entry.Qty * entry.Rate
		entry.GUID = helpers.GenerateGUID(entry.PartyName, entry.InvoiceNo, entry.Date)

		// --- XML Logic ---
		var taxXML string
		if entry.IGSTAmount > 0 {
			taxXML = fmt.Sprintf(constants.PurchaseVoucherIGSTTemplate, entry.IGSTAmount)
		} else {
			taxXML = fmt.Sprintf(constants.PurchaseVoucherCGSTSGSTTemplate, entry.CGSTAmount, entry.SGSTAmount)
		}

		// 🔥 NEW LOGIC: MULTI-COMPANY IMPORT
		// ==========================================

		// 1. Yahan apni dono companies ke EXACT naam likhein (Jo Tally me dikhte hain)
		targetCompanies := []string{
			"DS", // Parent company
			"DV", // Child company
		}

		// 2. Loop through companies
		for _, companyName := range targetCompanies {
			finalXML := fmt.Sprintf(constants.PurchaseVoucherXMLTemplate,
				companyName,                        // <SVCURRENTCOMPANY> (New)
				entry.Date,                         // DATE
				entry.Date,                         // REFERENCEDATE
				entry.GUID,                         // GUID
				helpers.EscapeXML(entry.InvoiceNo), // VOUCHERNUMBER
				helpers.EscapeXML(entry.InvoiceNo), // REFERENCE
				helpers.EscapeXML(entry.PartyName), // PARTYLEDGERNAME

				// --- Inventory Block ---
				helpers.EscapeXML(entry.ItemName), // STOCKITEMNAME
				entry.Rate,                        // RATE
				entry.Qty,                         // ACTUALQTY
				entry.Qty,                         // BILLEDQTY
				entry.Amount,                      // AMOUNT

				// --- Batch Allocation Block (REPEAT VALUES) ---
				entry.Amount, // Batch AMOUNT
				entry.Qty,    // Batch ACTUALQTY
				entry.Qty,    // Batch BILLEDQTY

				// --- Accounting Allocation Block (REPEAT VALUES) ---
				entry.Amount, // Accounting AMOUNT

				// --- Party Ledger ---
				helpers.EscapeXML(entry.PartyName), // LEDGERNAME
				entry.TotalBill,                    // AMOUNT (Total Positive)

				// --- Tax XML ---
				taxXML)

			fmt.Println("SENDING XML:", finalXML)

			// 6. Log: Sending to Tally
			tallyURL := "http://localhost:9000"
			resp, err := helpers.SendToTally(tallyURL, finalXML)

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
	}

	return fmt.Sprintf("Import Complete! Success: %d, Skipped: %d, Failed: %d", success, skipped, errors)
}
