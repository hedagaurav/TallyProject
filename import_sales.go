package main

import (
	"TallyProject/constants"
	"TallyProject/models"
	"fmt"
	"math"
	"strings"
)

func GenerateSalesXML(v models.SalesVoucher) string {
	var xmlBuilder strings.Builder

	// ---------------------------------------------------------
	// STEP 1: Main Body Populate karna
	// ---------------------------------------------------------
	// Values formatted for placeholders
	// Dhyan de: Sales me Credit amount Negative (-) hota hai
	itemAmt := -math.Abs(v.ItemAmount)
	salesAmt := -math.Abs(v.SalesAmount)
	unit := "Bags" // Ya variable se le lo

	bodyXML := fmt.Sprintf(constants.SALES_VOUCHER_BODY,
		v.VoucherDate,     // 1. Date
		v.VoucherNumber,   // 2. VNo
		v.PartyLedgerName, // 3. Party Name
		v.Narration,       // 4. Narration
		v.VoucherNumber,   // 5. FBID
		v.ItemName,        // 6. Item Name
		v.GodownName,      // 7. Godown
		itemAmt,           // 8. Amount (Negative)
		v.BilledQty, unit, // 9. Actual Qty
		v.BilledQty, unit, // 10. Billed Qty
		v.SalesLedgerName, // 11. Sales Ledger Name
		salesAmt,          // 12. Sales Amount (Negative)
	)

	xmlBuilder.WriteString(bodyXML)

	// ---------------------------------------------------------
	// STEP 2: Conditional Tax XML (Dynamic Append)
	// ---------------------------------------------------------

	// Check CGST
	if v.OutputCGSTAmount > 0 {
		taxAmt := -math.Abs(v.OutputCGSTAmount) // Negative
		taxXML := fmt.Sprintf(constants.SALES_TAX_LEDGER,
			v.OutputCGSTLedger,
			taxAmt,
		)
		xmlBuilder.WriteString(taxXML)
	}

	// Check SGST
	if v.OutputSGSTAmount > 0 {
		taxAmt := -math.Abs(v.OutputSGSTAmount) // Negative
		taxXML := fmt.Sprintf(constants.SALES_TAX_LEDGER,
			v.OutputSGSTLedger,
			taxAmt,
		)
		xmlBuilder.WriteString(taxXML)
	}

	// ---------------------------------------------------------
	// STEP 3: Party Ledger (Footer) Populate karna
	// ---------------------------------------------------------
	// Party hamesha Debit (Positive) hoti hai Sales me
	partyTotal := math.Abs(v.PartyAmount)

	footerXML := fmt.Sprintf(SALES_PARTY_LEDGER,
		v.PartyLedgerName, // 1. Party Name
		partyTotal,        // 2. Total Amount (Positive)
	)

	xmlBuilder.WriteString(footerXML)

	// Final XML Return
	return xmlBuilder.String()
}
