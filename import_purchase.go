package main

import (
	"TallyProject/constants"
	"TallyProject/helpers"
	"TallyProject/models"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

func readPurchaseRows(filePath string) ([][]string, error) {

	if filePath == "" {
		return nil, fmt.Errorf("file path empty")
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("no sheets found")
	}

	sheetName := sheetList[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	var validRows [][]string

	for i, row := range rows {

		if i < 2 {
			continue
		}

		if len(row) < 22 {
			continue
		}

		if row[0] != "D.V" {
			continue
		}

		validRows = append(validRows, row)
	}

	return validRows, nil
}

func parsePurchaseRow(row []string) (models.PurchaseEntry, error) {
	return mapRowToPurchaseEntry(row)
}

func detectCompanies(excelCompany string) []string {

	parent := "DS"
	excelCompany = strings.TrimSpace(excelCompany)

	// agar excel me company blank ho ya DS ho
	if excelCompany == "" || excelCompany == parent {
		return []string{parent}
	}

	// child company present
	return []string{parent, excelCompany}
}

func (a *App) ImportPurchaseVoucher(filePath string) string {
	// 1. Log: File Path Check
	fmt.Println("Step 1: File Path mile ->", filePath)

	if filePath == "" {
		return "Please select a file first!"
	}

	rows, err := readPurchaseRows(filePath)
	if err != nil {
		fmt.Println("Error reading Excel:", err)
		return "Error: Could not read Excel file"
	}

	fmt.Printf("Valid Purchase Rows -> %d\n", len(rows))

	success := 0
	skipped := 0
	errors := 0

	// Loop Starts
	for i, row := range rows {
		fmt.Printf("Processing Row %d...\n", i+1)

		entry, err := parsePurchaseRow(row)
		if err != nil {
			fmt.Printf("❌ Row %d Parse Error: %v\n", i+1, err)
			errors++
			continue
		}

		// NEW LOGIC: MULTI-COMPANY IMPORT
		companies := detectCompanies(row[3])

		// 2. Loop through companies
		for _, companyName := range companies {
			finalXML := buildPurchaseXML(entry, companyName)

			fmt.Println("SENDING XML:", finalXML)

			// 6. Log: Sending to Tally
			// resp, err := helpers.SendToTally(constants.TallyURL, finalXML)

			// if err != nil {
			// 	fmt.Printf("Row %d Network Error: %v\n", i+1, err)
			// 	errors++
			// } else if strings.Contains(resp, "<CREATED>1</CREATED>") {
			// 	fmt.Printf("Row %d Success!\n", i+1)
			// 	success++
			// } else {
			// 	// Duplicate GUID error ya koi aur logic error check karne ke liye:
			// 	fmt.Printf("Row %d Tally Rejected: %s\n", i+1, resp)
			// 	errors++
			// }
		}
	}

	return fmt.Sprintf("Import Complete! Success: %d, Skipped: %d, Failed: %d", success, skipped, errors)
}

func buildPurchaseXML(entry models.PurchaseEntry, companyName string) string {

	// --- TAX XML ---
	var taxXML string
	if entry.IGSTAmount > 0 {
		taxXML = fmt.Sprintf(constants.PurchaseVoucherIGSTTemplate, entry.IGSTAmount)
	} else {
		taxXML = fmt.Sprintf(
			constants.PurchaseVoucherCGSTSGSTTemplate,
			entry.CGSTAmount,
			entry.SGSTAmount,
		)
	}

	// --- FINAL XML ---
	finalXML := fmt.Sprintf(constants.PurchaseVoucherXMLTemplate,
		companyName,                        // SVCURRENTCOMPANY
		entry.Date,                         // DATE
		entry.Date,                         // REFERENCEDATE
		entry.GUID,                         // GUID
		helpers.EscapeXML(entry.InvoiceNo), // VOUCHERNUMBER
		helpers.EscapeXML(entry.InvoiceNo), // REFERENCE
		helpers.EscapeXML(entry.PartyName), // PARTYLEDGERNAME

		// Inventory
		helpers.EscapeXML(entry.ItemName),
		entry.Rate,
		entry.Qty,
		entry.Qty,
		entry.Amount,

		// Batch
		entry.Amount,
		entry.Qty,
		entry.Qty,

		// Accounting
		entry.Amount,

		// Party ledger
		helpers.EscapeXML(entry.PartyName),
		entry.TotalBill,

		// Tax
		taxXML,
	)

	return strings.TrimSpace(finalXML)
}

func mapRowToPurchaseEntry(row []string) (models.PurchaseEntry, error) {

	var entry models.PurchaseEntry

	// Date
	excelFormattedDate, err := helpers.ParseDateSmart(row[1])
	if err != nil {
		return entry, fmt.Errorf("invalid date: %s", row[1])
	}

	entry.Date = excelFormattedDate
	entry.PartyName = row[2]
	entry.InvoiceNo = row[4]
	entry.ItemName = row[5]
	entry.Qty = helpers.ParseFloat(row[7])
	entry.Rate = helpers.ParseFloat(row[8])

	entry.IGSTAmount = helpers.ParseFloat(row[12])
	entry.CGSTAmount = helpers.ParseFloat(row[13])
	entry.SGSTAmount = helpers.ParseFloat(row[14])
	entry.TotalBill = helpers.ParseFloat(row[21])

	// Derived
	entry.Amount = entry.Qty * entry.Rate
	entry.GUID = helpers.GenerateGUID(entry.PartyName, entry.InvoiceNo, entry.Date)

	return entry, nil
}
