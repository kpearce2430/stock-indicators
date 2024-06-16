package model

import "github.com/sirupsen/logrus"

type SymbolType struct {
	Description string `json:"description,omitempty"`
	Symbol      string `json:"symbol,omitempty"`
	Type        string `json:"type,omitempty"`
}

var SymbolTypeMap map[string]string

// TODO:  Use CDB or Postgres to store.
func init() {
	logrus.Info("loading symbol type map")
	SymbolTypes := []SymbolType{
		{
			Description: "Alphabet Inc. - Class C Capital Stock",
			Symbol:      "GOOG",
			Type:        "Stock",
		},
		{Description: "USAA Mutual Fds Tr Nasdaq 100 Index Fund",
			Symbol: "USNQX",
			Type:   "Mutual Fund",
		},
		{
			Description: "AT&T (T)",
			Symbol:      "T",
			Type:        "Stock",
		},
		{
			Description: "APPLE INC COM",
			Symbol:      "AAPL",
			Type:        "Stock",
		},
		{
			Description: "Fidelity Countra Fund",
			Symbol:      "FCNTX",
			Type:        "Mutual Fund",
		},
		{
			Description: "EMERSON ELEC CO COM",
			Symbol:      "EMR",
			Type:        "Stock",
		},
		{
			Description: "Home Depot",
			Symbol:      "HD",
			Type:        "Stock",
		},
		{
			Description: "3M Corp (MMM)",
			Symbol:      "MMM",
			Type:        "Stock",
		},
		{
			Description: "STATE STREET S&P 500 INDEX N",
			Symbol:      "SVSPX",
			Type:        "Mutual Fund",
		},
		{
			Description: "Large Cap Index",
			Symbol:      "HDLCAP",
			Type:        "Mutual Fund",
		},
		{
			Description: "Microsoft Corp",
			Symbol:      "MSFT",
			Type:        "Stock",
		},
		{
			Description: "Small-Mid Cap Index",
			Symbol:      "HDMCI",
			Type:        "Mutual Fund",
		},
		{
			Description: "NB Genesis Trust",
			Symbol:      "NBGEX",
			Type:        "Mutual Fund",
		},
		{
			Description: "Jensen Quality Growth Fund Cl J",
			Symbol:      "JENSX",
			Type:        "Mutual Fund",
		},
		{
			Description: "Balanced",
			Symbol:      "HDBAL",
			Type:        "Mutual Fund",
		},
		{
			Description: "PFIZER INC COM",
			Symbol:      "PFE",
			Type:        "Stock",
		},
		{
			Description: "LifePath 2030 Portfolio",
			Symbol:      "HD2030",
			Type:        "Mutual Fund",
		},
		{
			Description: "International Equity Index",
			Symbol:      "HDIEI",
			Type:        "Mutual Fund",
		},
		{
			Description: "Union Pacific Railroad",
			Symbol:      "UNP",
			Type:        "Stock",
		},
		{
			Description: "FIDELITY MULTI-ASSET INCOME FUND",
			Symbol:      "FMSDX",
			Type:        "Mutual Fund",
		},
		{
			Description: "JOHNSON & JOHNSON COM",
			Symbol:      "JNJ",
			Type:        "Stock",
		},
		{
			Description: "CSX Corp",
			Symbol:      "CSX",
			Type:        "Stock",
		},
		{
			Description: "SOUTHERN CO",
			Symbol:      "SO",
			Type:        "Other",
		},
		{
			Description: "FIDELITY CAPITAL & INCOME",
			Symbol:      "FAGIX",
			Type:        "Mutual Fund",
		},
		{
			Description: "USAA Mutual Fds Tr Income Fund",
			Symbol:      "USAIX",
			Type:        "Mutual Fund",
		},
		{
			Description: "BRISTOL MYERS SQUIBB CO COM",
			Symbol:      "BMY",
			Type:        "Stock",
		},
		{
			Description: "BANK OZK LITTLE R 4.7%24CD FDIC INS DUE 08/09/24US",
			Symbol:      "06418CPW0",
			Type:        "Bond",
		},
		{
			Description: "BANK OZK CD MTHLY",
			Symbol:      "06418CJZ0",
			Type:        "Bond",
		},
		{
			Description: "680061KD9 OLD NATL BK EVANSVILLE IND CD 5.30000% 08/22/2024",
			Symbol:      "680061KD9",
			Type:        "Bond",
		},
		{
			Description: "GOLDMAN SACHS BA 4.75%25CD FDIC INS DUE 01/16/25US",
			Symbol:      "38150VRQ4",
			Type:        "Bond",
		},
		{
			Description: "06418CMA1 BANK OZK LITTLE ROCK ARK CD 5.40000% 10/04/2024",
			Symbol:      "06418CMA1",
			Type:        "Bond",
		},
		{
			Description: "GOLDMAN SACHS BAN 5.5%24CD FDIC INS DUE 10/15/24US",
			Symbol:      "38150VNM7",
			Type:        "Bond",
		},
		{
			Description: "GOLDMAN SACHS BA 5.55%24CD FDIC INS DUE 10/09/24US",
			Symbol:      "38150VMY2",
			Type:        "Bond",
		},
		{
			Description: "JPMORGAN CHASE & 5.65%24CD FDIC INS DUE 11/05/24US",
			Symbol:      "46656MTK7",
			Type:        "Bond",
		},
		{
			Description: "JPMORGAN CHASE BANK NATIONAL A CD M/W CL",
			Symbol:      "46656MKG5",
			Type:        "Bond",
		},
		{
			Description: "46656MWW7 JPMORGAN CHASE BK N A CD 5.50000% 12/06/2024",
			Symbol:      "46656MWW7",
			Type:        "Bond",
		},
		{
			Description: "Mondelez International Inc",
			Symbol:      "MDLZ",
			Type:        "Stock",
		},
		{
			Description: "Norfolk Southern",
			Symbol:      "NSC",
			Type:        "Stock",
		},
		{
			Description: "Altria Group",
			Symbol:      "MO",
			Type:        "Stock",
		},
		{
			Description: "PROCTER & GAMBLE CO COM",
			Symbol:      "PG",
			Type:        "Stock",
		},
		{
			Description: "Coke",
			Symbol:      "KO",
			Type:        "Stock",
		},
		{
			Description: "PEPSICO INC",
			Symbol:      "PEP",
			Type:        "Stock",
		},
		{
			Description: "PHILIP MORRIS CO INC",
			Symbol:      "PM",
			Type:        "Stock",
		},
	}
	SymbolTypeMap = make(map[string]string)
	for _, s := range SymbolTypes {
		SymbolTypeMap[s.Symbol] = s.Type
	}
}
