package worksheets

import "fmt"

func (w *WorkSheet) PortfolioValueSheet(worksheetName string) error {
	//
	_, err := w.File.NewSheet(worksheetName)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return err
	}

	return nil
}
