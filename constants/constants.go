package constants

// SECURITY CONSTANT (Isse koi guess nahi kar payega)
// Isse apne hisaab se change kar lena
const AppSecret = "Gaurav_Tally_Secret_2026_Key"
const TsallyURL = "http://localhost:9000"

const (
	// Tally ko ye format pasand hai
	TallyDateFormat = "20060102"

	// Excel se ye format aa raha hai (DD-MM-YYYY)
	ExcelDateFormat = "02-01-2006"

	// Purchase Tax Templates (Input Tax, Debit, Negative Amount)
	// %f jahan amount aayega
	PurchaseIGSTTemplate = `
    <LEDGERENTRIES.LIST>
        <LEDGERNAME>Input IGST</LEDGERNAME>
        <ISDEEMEDPOSITIVE>Yes</ISDEEMEDPOSITIVE>
        <AMOUNT>-%f</AMOUNT>
    </LEDGERENTRIES.LIST>`

	// Local Tax (CGST + SGST combined)
	// Isme 2 baar %f aayega (ek CGST ke liye, ek SGST ke liye)
	PurchaseCGSTSGSTTemplate = `
    <LEDGERENTRIES.LIST>
        <LEDGERNAME>Input CGST</LEDGERNAME>
        <ISDEEMEDPOSITIVE>Yes</ISDEEMEDPOSITIVE>
        <AMOUNT>-%f</AMOUNT>
    </LEDGERENTRIES.LIST>
    <LEDGERENTRIES.LIST>
        <LEDGERNAME>Input SGST</LEDGERNAME>
        <ISDEEMEDPOSITIVE>Yes</ISDEEMEDPOSITIVE>
        <AMOUNT>-%f</AMOUNT>
    </LEDGERENTRIES.LIST>`

	PurchaseXMLTemplate = `
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
                    <VOUCHER VCHTYPE="Purchase" ACTION="Create" OBJVIEW="Invoice Voucher View">
                        <ISINVOICE>Yes</ISINVOICE>
                        
                        <DATE>%s</DATE>
                        <REFERENCEDATE>%s</REFERENCEDATE>
                        <GUID>%s</GUID>
                        <VOUCHERTYPENAME>Purchase</VOUCHERTYPENAME>
                        <VOUCHERNUMBER>%s</VOUCHERNUMBER>
                        <REFERENCE>%s</REFERENCE>
                        <PARTYLEDGERNAME>%s</PARTYLEDGERNAME>
                        <FBTPAYMENTTYPE>Default</FBTPAYMENTTYPE>
                        <PERSISTEDVIEW>Invoice Voucher View</PERSISTEDVIEW> <ALLINVENTORYENTRIES.LIST>
                            <STOCKITEMNAME>%s</STOCKITEMNAME>
                            <ISDEEMEDPOSITIVE>Yes</ISDEEMEDPOSITIVE>
                            <ISLASTDEEMEDPOSITIVE>Yes</ISLASTDEEMEDPOSITIVE> <ISINWARD>Yes</ISINWARD> <RATE>%f/Bag</RATE>
                            <ACTUALQTY> %f Bag</ACTUALQTY>
                            <BILLEDQTY> %f Bag</BILLEDQTY>
                            <AMOUNT>-%f</AMOUNT> <BATCHALLOCATIONS.LIST>
                                <GODOWNNAME>Main Location</GODOWNNAME> <BATCHNAME>Primary Batch</BATCHNAME>
                                <AMOUNT>-%f</AMOUNT> <ACTUALQTY> %f Bag</ACTUALQTY>
                                <BILLEDQTY> %f Bag</BILLEDQTY>
                            </BATCHALLOCATIONS.LIST>

                            <ACCOUNTINGALLOCATIONS.LIST>
                                <LEDGERNAME>Purchase Account</LEDGERNAME>
                                <ISDEEMEDPOSITIVE>Yes</ISDEEMEDPOSITIVE>
                                <AMOUNT>-%f</AMOUNT> </ACCOUNTINGALLOCATIONS.LIST>
                        </ALLINVENTORYENTRIES.LIST>

                        <LEDGERENTRIES.LIST>
                            <LEDGERNAME>%s</LEDGERNAME>
                            <ISDEEMEDPOSITIVE>No</ISDEEMEDPOSITIVE>
                            <ISLASTDEEMEDPOSITIVE>No</ISLASTDEEMEDPOSITIVE>
                            <AMOUNT>%f</AMOUNT> </LEDGERENTRIES.LIST>

                        %s 

                    </VOUCHER>
                </TALLYMESSAGE>
            </REQUESTDATA>
        </IMPORTDATA>
    </BODY>
	</ENVELOPE>`
)
