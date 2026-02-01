package models

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

// Excel Data store karne ke liye structure
type PurchaseEntry struct {
	GUID      string // Unique ID (Duplicate rokne ke liye)
	Date      string // Tally Format: YYYYMMDD
	PartyName string // RCCPL PVT LTD
	InvoiceNo string // 2711142290
	ItemName  string // MP Birla Perfect Plus
	Qty       float64
	Rate      float64
	Amount    float64 // Item Amount (Qty * Rate)

	// Tax Info
	IGSTAmount float64
	CGSTAmount float64
	SGSTAmount float64
	TotalBill  float64
}

// 1. Excel se Raw Data padhne ke liye Struct
type RawSalesRow struct {
	Date        string  // Excel: Date
	CompanyCode string  // Excel: New Column (D.V / D.S)
	PartyName   string  // Excel: Bill Name
	VoucherNo   string  // Excel: B.No
	ChallanNo   string  // Excel: Ch. No (Green Section)
	ItemName    string  // Excel: Brand
	Qty         float64 // Excel: Qty
	Rate        float64 // Excel: Rate
	GrossAmount float64 // Excel: G. Amt (Taxable)
	CGST        float64 // Excel: CGST
	SGST        float64 // Excel: SGST
	TotalAmount float64 // Excel: Total Amt (Party Debit amount)
}

// 2. Tally XML generate karne ke liye Final Struct
type SalesVoucher struct {
	VoucherType   string
	VoucherDate   string // YYYYMMDD format
	VoucherNumber string
	Narration     string

	// Party Details (Debit / Positive)
	PartyLedgerName string
	PartyAmount     float64

	// Inventory Details (Credit / Negative)
	ItemName   string
	GodownName string
	BilledQty  float64
	Rate       float64
	ItemAmount float64 // Negative Value

	// Accounting Allocations (Credit / Negative)
	SalesLedgerName string
	SalesAmount     float64 // Negative Value

	// Tax Details (Credit / Negative)
	OutputCGSTLedger string
	OutputCGSTAmount float64 // Negative Value

	OutputSGSTLedger string
	OutputSGSTAmount float64 // Negative Value

	// Meta Data for internal logic
	TargetCompany string
}
