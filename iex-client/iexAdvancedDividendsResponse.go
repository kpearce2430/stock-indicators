package iex_client

import "github.com/sirupsen/logrus"

type IEXAdvancedDividendsResponse struct {
	AdrFee                int     `json:"adrFee,omitempty"`
	Amount                float64 `json:"amount,omitempty"`
	AnnounceDate          string  `json:"announceDate,omitempty"`
	CountryCode           string  `json:"countryCode,omitempty"`
	Coupon                int     `json:"coupon,omitempty"`
	Created               string  `json:"created,omitempty"`
	Currency              string  `json:"currency,omitempty"`
	DeclaredCurrencyCD    string  `json:"declaredCurrencyCD,omitempty"`
	DeclaredDate          string  `json:"declaredDate,omitempty"`
	DeclaredGrossAmount   int     `json:"declaredGrossAmount,omitempty"`
	Description           string  `json:"description,omitempty"`
	ExDate                string  `json:"exDate,omitempty"`
	Figi                  string  `json:"figi,omitempty"`
	FiscalYearEndDate     string  `json:"fiscalYearEndDate,omitempty"`
	Flag                  string  `json:"flag,omitempty"`
	Frequency             string  `json:"frequency,omitempty"`
	FromFactor            int     `json:"fromFactor,omitempty"`
	FxDate                string  `json:"fxDate,omitempty"`
	GrossAmount           float64 `json:"grossAmount,omitempty"`
	InstallmentPayDate    any     `json:"installmentPayDate,omitempty"`
	IsApproximate         any     `json:"isApproximate,omitempty"`
	IsCapitalGains        any     `json:"isCapitalGains,omitempty"`
	IsDAP                 any     `json:"isDAP,omitempty"`
	IsNetInvestmentIncome any     `json:"isNetInvestmentIncome,omitempty"`
	LastUpdated           string  `json:"lastUpdated,omitempty"`
	Marker                string  `json:"marker,omitempty"`
	NetAmount             float64 `json:"netAmount,omitempty"`
	Notes                 string  `json:"notes,omitempty"`
	OptionalElectionDate  string  `json:"optionalElectionDate,omitempty"`
	ParValue              float64 `json:"parValue,omitempty"`
	ParValueCurrency      string  `json:"parValueCurrency,omitempty"`
	PaymentDate           string  `json:"paymentDate,omitempty"`
	PeriodEndDate         string  `json:"periodEndDate,omitempty"`
	RecordDate            string  `json:"recordDate,omitempty"`
	Refid                 string  `json:"refid,omitempty"`
	RegistrationDate      string  `json:"registrationDate,omitempty"`
	SecondExDate          string  `json:"secondExDate,omitempty"`
	SecondPaymentDate     string  `json:"secondPaymentDate,omitempty"`
	SecurityType          string  `json:"securityType,omitempty"`
	Symbol                string  `json:"symbol,omitempty"`
	TaxRate               int     `json:"taxRate,omitempty"`
	ToDate                string  `json:"toDate,omitempty"`
	ToFactor              int     `json:"toFactor,omitempty"`
	ID                    string  `json:"id,omitempty"`
	Key                   string  `json:"key,omitempty"`
	Subkey                string  `json:"subkey,omitempty"`
	Date                  int64   `json:"date,omitempty"`
	Updated               int64   `json:"updated,omitempty"`
}

func (d IEXAdvancedDividendsResponse) GetSymbol() string {
	return d.Symbol
}

func (d IEXAdvancedDividendsResponse) YearlyDividend() float64 {

	switch d.Frequency {

	case "quarterly":
		return d.Amount * 4.0
	case "yearly":
		return d.Amount
	default:
		logrus.Info("Unhandled Frequency:", d.Frequency)
		return d.Amount
	}
	return -1.00
}
