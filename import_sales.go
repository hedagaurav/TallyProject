package main

import (
	"TallyProject/constants"
	"TallyProject/helpers"
	"TallyProject/models"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"
)

func ParseExcelRows(excelRows [][]string) {

	for i, row := range excelRows {
		// Skip Header Rows (Row 1 and 2)
		if i < 3 {
			continue
		}

		// Validation: Ensure row has enough columns
		// Hum last column 'Total Amt' (Col BD -> Index 55) access kar rahe hain,
		// to length kam se kam 56 honi chahiye.
		if len(row) < 56 {
			// Optional: Log skipped rows for debugging
			// fmt.Printf("Row %d skipped: Insufficient columns (%d)\n", i+1, len(row))
			continue
		}
		
		var currentRow models.SalesVoucherDTO

		// DATE PARSING (Bill Date - Column AS -> Index 44)
		excelFormattedDate, err := helpers.ParseDateSmart(row[44])
		if err != nil {
			fmt.Printf("❌ Row %d Date Error: '%s'\n", i+1, row[44])
			errors++
			continue
		}

		currentRow.Date = excelFormattedDate

		// COMPANY NAME (Column AB -> Index 27)
		currentRow.CompanyCode = strings.TrimSpace(row[27])

					// --- BILL DETAILS (RED SECTION) ---
			// Voucher No: Column AT -> Index 45 ("B.No.")
			currentRow.VoucherNumber = row[45]

			// Party Name: Column AU -> Index 46 ("Bill Name")
			currentRow.PartyName = row[46]

			// Billed Qty: Column AV -> Index 47 ("Qty")
			currentRow.BilledQty = helpers.ParseFloat(row[47])

			// Rate: Column AW -> Index 48 ("Rate")
			currentRow.ItemRate = helpers.ParseFloat(row[48])

			// Taxable Amount: Column AZ -> Index 51 ("G. Amt")
			ItemAmount:  helpers.ParseFloat(row[51]),
			SalesAmount: helpers.ParseFloat(row[51]), // Usually same as ItemAmount

			// Tax Data: Column BA -> Index 52 ("CGST")
			OutputCGSTAmount: helpers.ParseFloat(row[52]),

			// Tax Data: Column BB -> Index 53 ("SGST")
			OutputSGSTAmount: helpers.ParseFloat(row[53]),

			// Total Bill: Column BD -> Index 55 ("Total Amt")
			PartyAmount: helpers.ParseFloat(row[55]),

			// --- MATERIAL DETAILS (GREEN SECTION) ---
			// Challan No: Column AF -> Index 31 ("Ch. No.")
			// Used for Narration
			Narration: fmt.Sprintf("Against Ch No: %s", row[31]),

			// Item Name: Column AG -> Index 32 ("Brand")
			ItemName: row[32],

			// Actual Qty: Column AH -> Index 33 ("QTY")
			ActualQty: helpers.ParseFloat(row[33]),

			// IGST (Future Proofing)
			// Filhal Data me column nahi hai, isliye 0.0 set kar rahe hain.
			// Agar future me Column 'BE' me aata hai to yahan 'helpers.ParseFloat(row[56])' kar dena.
			OutputIGSTAmount: 0.0,
	}
}

func GenerateXMLFromData() string {
	var XMLstring string

	XMLstring = ""

	return XMLstring
}

func (a *App) ImportSalesVoucher(filePath string) string {
	// 1. File Opening & Setup
	fmt.Println("Step 1: Sales File Path ->", filePath)
	if filePath == "" {
		return "Please select a file first!"
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return "Error: Could not open Excel file"
	}
	defer f.Close()

	// 2. Sheet Selection
	sheetList := f.GetSheetList()
	sheetName := "April-25" // Default
	if len(sheetList) > 0 {
		sheetName = sheetList[0] // Pick first sheet dynamically
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return "Error: Sheet name not found"
	}

	fmt.Printf("Step 3: Total Rows Found -> %d\n", len(rows))

	ParseExcelRows(rows)
	os.Exit(1)
	success, skipped, errors := 0, 0, 0

	// 3. Row Processing Loop

	for i, row := range rows {
		// Skip Header Rows (Row 1 and 2)
		if i < 3 {
			continue
		}

		// Validation: Ensure row has enough columns
		// Hum last column 'Total Amt' (Col BD -> Index 55) access kar rahe hain,
		// to length kam se kam 56 honi chahiye.
		if len(row) < 56 {
			// Optional: Log skipped rows for debugging
			// fmt.Printf("Row %d skipped: Insufficient columns (%d)\n", i+1, len(row))
			continue
		}

		// --- COLUMN MAPPING LOGIC START ---

		// 1. COMPANY NAME (Column AB -> Index 27)
		rowCompanyCode := strings.TrimSpace(row[27])
		if rowCompanyCode == "" {
			skipped++
			continue
		}

		// Company Selection Logic
		var targetCompany string
		targetCompany = "DV"

		fmt.Printf("Processing Row %d for %s...\n", i+1, targetCompany)

		// 2. DATE PARSING (Bill Date - Column AS -> Index 44)
		excelFormattedDate, err := helpers.ParseDateSmart(row[44])
		if err != nil {
			fmt.Printf("❌ Row %d Date Error: '%s'\n", i+1, row[44])
			errors++
			continue
		}

		fmt.Println("date = ", row[44])
		fmt.Println("date fmtt= ", excelFormattedDate)
		// 3. FILLING THE DTO
		entry := models.SalesVoucherDTO{
			// Metadata
			VoucherType:   "Sales",
			TargetCompany: targetCompany,
			VoucherDate:   excelFormattedDate,

			// --- BILL DETAILS (RED SECTION) ---
			// Voucher No: Column AT -> Index 45 ("B.No.")
			VoucherNumber: row[45],

			// Party Name: Column AU -> Index 46 ("Bill Name")
			PartyName: row[46],

			// Billed Qty: Column AV -> Index 47 ("Qty")
			BilledQty: helpers.ParseFloat(row[47]),

			// Rate: Column AW -> Index 48 ("Rate")
			ItemRate: helpers.ParseFloat(row[48]),

			// Taxable Amount: Column AZ -> Index 51 ("G. Amt")
			ItemAmount:  helpers.ParseFloat(row[51]),
			SalesAmount: helpers.ParseFloat(row[51]), // Usually same as ItemAmount

			// Tax Data: Column BA -> Index 52 ("CGST")
			OutputCGSTAmount: helpers.ParseFloat(row[52]),

			// Tax Data: Column BB -> Index 53 ("SGST")
			OutputSGSTAmount: helpers.ParseFloat(row[53]),

			// Total Bill: Column BD -> Index 55 ("Total Amt")
			PartyAmount: helpers.ParseFloat(row[55]),

			// --- MATERIAL DETAILS (GREEN SECTION) ---
			// Challan No: Column AF -> Index 31 ("Ch. No.")
			// Used for Narration
			Narration: fmt.Sprintf("Against Ch No: %s", row[31]),

			// Item Name: Column AG -> Index 32 ("Brand")
			ItemName: row[32],

			// Actual Qty: Column AH -> Index 33 ("QTY")
			ActualQty: helpers.ParseFloat(row[33]),

			// IGST (Future Proofing)
			// Filhal Data me column nahi hai, isliye 0.0 set kar rahe hain.
			// Agar future me Column 'BE' me aata hai to yahan 'helpers.ParseFloat(row[56])' kar dena.
			OutputIGSTAmount: 0.0,
		}

		// --- HARDCODED DEFAULTS (As per previous instructions) ---
		entry.Unit = "Bags"
		// entry.GodownName = "Main Location"
		entry.SalesLedger = "Sales Account"
		entry.OutputCGSTLedger = "Output CGST"
		entry.OutputSGSTLedger = "Output SGST"
		entry.OutputIGSTLedger = "Output IGST"

		fmt.Println("struct data:", entry)
		// --- XML GENERATION & SENDING ---

		// Sign Conversions (Credit = Negative, Debit = Positive)
		itemAmt := -math.Abs(entry.ItemAmount)
		salesAmt := -math.Abs(entry.SalesAmount)
		cgstAmt := -math.Abs(entry.OutputCGSTAmount)
		sgstAmt := -math.Abs(entry.OutputSGSTAmount)
		igstAmt := -math.Abs(entry.OutputIGSTAmount)
		partyAmt := math.Abs(entry.PartyAmount)

		// Dynamic Tax XML Construction
		var taxXML string

		// IGST Check
		if entry.OutputIGSTAmount > 0 {
			taxXML += fmt.Sprintf(constants.SalesTaxTemplate, entry.OutputIGSTLedger, igstAmt)
		}
		// CGST Check
		if entry.OutputCGSTAmount > 0 {
			taxXML += fmt.Sprintf(constants.SalesTaxTemplate, entry.OutputCGSTLedger, cgstAmt)
		}
		// SGST Check
		if entry.OutputSGSTAmount > 0 {
			taxXML += fmt.Sprintf(constants.SalesTaxTemplate, entry.OutputSGSTLedger, sgstAmt)
		}

		// Final XML Formatting
		finalXML := fmt.Sprintf(constants.SalesVoucherXMLTemplate,
			targetCompany, // <SVCURRENTCOMPANY>
			// entry.VoucherDate,                  // DATE
			entry.VoucherNumber,                // VOUCHERNUMBER
			helpers.EscapeXML(entry.PartyName), // PARTYLEDGERNAME
			helpers.EscapeXML(entry.Narration), // NARRATION
			entry.VoucherNumber,                // FBID (Unique ID)

			// Inventory Details
			helpers.EscapeXML(entry.ItemName), // STOCKITEMNAME
			// entry.GodownName,                  // GODOWNNAME
			itemAmt,                     // AMOUNT (Negative)
			entry.ActualQty, entry.Unit, // ACTUALQTY (Green Section Data)
			entry.BilledQty, entry.Unit, // BILLEDQTY (Red Section Data)

			// Sales Ledger Allocation
			entry.SalesLedger, // LEDGERNAME
			salesAmt,          // AMOUNT (Negative)

			// Tax Block
			taxXML,

			// Party Ledger (Debit)
			helpers.EscapeXML(entry.PartyName), // LEDGERNAME
			partyAmt,                           // AMOUNT (Positive)
		)

		// Final Clean Up before sending
		finalXML = strings.TrimSpace(finalXML)

		fmt.Println("SALES XML = ", finalXML)
		fmt.Println("SENDING SALES XML...", entry.VoucherNumber)

		// 4. Send to Tally
		tallyURL := "http://localhost:9000"
		resp, err := helpers.SendToTally(tallyURL, finalXML)

		if err != nil {
			fmt.Printf("❌ Row %d Network Error: %v\n", i+1, err)
			errors++
		} else if strings.Contains(resp, "<CREATED>1</CREATED>") {
			fmt.Printf("✅ Row %d Success! Created in %s\n", i+1, targetCompany)
			success++
		} else {
			fmt.Printf("❌ Row %d Tally Rejected: %s\n", i+1, resp)
			errors++
		}
	}

	return fmt.Sprintf("Import Complete! Success: %d, Skipped: %d, Failed: %d", success, skipped, errors)
}
