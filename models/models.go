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
