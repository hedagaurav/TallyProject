package helpers

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	constant "TallyProject/constants"
)

func ParseFloat(s string) float64 {
	// Hidden characters aur spaces saaf karne ke liye
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "") // Agar 1,12,000 jaisa format ho

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return val
}

// Updated GUID Generator (Ab ye Date bhi leta hai)
func GenerateGUID(partyName, invoiceNo, date string) string {
	// Logic: Party + Invoice + Date
	// Date add karne se agar agle saal same Invoice No aaya to duplicate nahi banega
	data := partyName + "-" + invoiceNo + "-" + date

	// MD5 Hash banao
	h := md5.New()
	io.WriteString(h, data)

	// Tally format me ID return karo
	return fmt.Sprintf("%x-0000-0000-0000-000000000000", h.Sum(nil))
}

// Smart Date Parser: Alag alag formats try karega
func ParseDateSmart(dateStr string) (string, error) {
	// List of formats to try
	formats := []string{
		"02-01-2006", // DD-MM-YYYY (User ka format)
		"02-01-06",   // DD-MM-YY   (2-digit Year)
		"2006-01-02", // YYYY-MM-DD (ISO)
	}

	for _, format := range formats {
		parsedDate, err := time.Parse(format, dateStr)
		if err == nil {
			// Agar success hua, to Tally format (YYYYMMDD) return karo
			return parsedDate.Format(constant.TallyDateFormat), nil
		}
	}

	return "", fmt.Errorf("unknown date format")
}

func EscapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// Helper Function: Jo XML ko Tally tak lekar jayega
// Return type ab (string, error) hai
func SendToTally(url, xmlData string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(xmlData)))
	if err != nil {
		return "", err // Error aane par empty string bhejo
	}
	req.Header.Set("Content-Type", "text/xml")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Response Body ko Padhna (Read) Zaroori hai
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Byte array ko String banakar wapas karo
	return string(body), nil
}
