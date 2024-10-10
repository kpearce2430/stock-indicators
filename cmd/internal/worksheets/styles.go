package worksheets

import "github.com/xuri/excelize/v2"

type Styles struct {
	Header     int
	Accounting []int
	Currency   []int
	Date       []int
	General    []int
	Number     []int
	Percent    []int
	Text       []int
}

var (
	currFmt        = "_($#,##0.00_);[Red]($#,##0.00)"
	dateFmt        = "mm/dd/yyy"
	numFmt         = "#,##0.00#"
	pctFmt         = "0.00 pct"
	defaultColours = []string{"#FFFFFF", "#D9E2F3"}
)

func DefaultStyles(f *excelize.File) (*Styles, error) {
	s := Styles{}
	// Header
	header, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4674C1"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Font: &excelize.Font{
			Size: 14.0,
			Bold: true,
		},
		// NumFmt: 0,
	})

	if err != nil {
		return nil, err
	}
	s.Header = header

	for _, c := range defaultColours {

		acctStyle, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{c}, Pattern: 1},
			Font: &excelize.Font{
				Size: 14.0,
			},
			NumFmt: 44,
		})
		if err != nil {
			return nil, err
		}
		s.Accounting = append(s.Accounting, acctStyle)

		currStyle, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{c}, Pattern: 1},
			Font: &excelize.Font{
				Size: 14.0,
			},
			CustomNumFmt: &currFmt,
		})
		if err != nil {
			return nil, err
		}
		s.Currency = append(s.Currency, currStyle)

		dateStyle, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{c}, Pattern: 1},
			Font: &excelize.Font{
				Size: 14.0,
			},
			CustomNumFmt: &dateFmt,
		})
		if err != nil {
			return nil, err
		}
		s.Date = append(s.Date, dateStyle)

		genStyle, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{c}, Pattern: 1},
			Font: &excelize.Font{
				Size: 14.0,
			},
			NumFmt: 0,
		})
		if err != nil {
			return nil, err
		}
		s.General = append(s.General, genStyle)

		numStyle, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{c}, Pattern: 1},
			Font: &excelize.Font{
				Size: 14.0,
			},
			CustomNumFmt: &numFmt,
		})
		if err != nil {
			return nil, err
		}
		s.Number = append(s.Number, numStyle)

		// Percent
		pctStyle, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{c}, Pattern: 1},
			Font: &excelize.Font{
				Size: 14.0,
			},
			// CustomNumFmt: &pctFmt,
			NumFmt: 10,
		})
		if err != nil {
			return nil, err
		}
		s.Percent = append(s.Percent, pctStyle)

		// Text
		style, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{c}, Pattern: 1},
			Font: &excelize.Font{
				Size: 14.0,
			},
		})
		if err != nil {
			return nil, err
		}
		s.Text = append(s.Text, style)

	}
	return &s, nil
}

func (s *Styles) AccountingStyle(row int) int {
	return s.Accounting[row%len(s.Accounting)]
}

func (s *Styles) CurrencyStyle(row int) int {
	return s.Currency[row%len(s.Currency)]
}

func (s *Styles) DateStyle(row int) int {
	return s.Date[row%len(s.Date)]
}

func (s *Styles) GeneralStyle(row int) int {
	return s.General[row%len(s.Date)]
}

func (s *Styles) NumberStyle(row int) int {
	return s.Number[row%len(s.Number)]
}

func (s *Styles) PercentStyle(row int) int {
	return s.Percent[row%len(s.Percent)]
}

func (s *Styles) TextStyle(row int) int {
	return s.Text[row%len(s.Text)]
}
