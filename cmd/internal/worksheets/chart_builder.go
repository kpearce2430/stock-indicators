package worksheets

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type ChartBuilder struct {
	WorksheetName string
	Title         string
	Cell          string
	Values        []ChartBuilderSeries
	Categories    []ChartBuilderSeries
	Type          excelize.ChartType
	Height        uint
	Width         uint
	VaryColors    bool
	xFont         *excelize.Font
	yFont         *excelize.Font
	xReverse      bool
	yReverse      bool
}

type ChartBuilderSeries struct {
	ColumnStart string
	RowStart    int
	ColumnEnd   string
	RowEnd      int
}

func (c *ChartBuilder) AddValueSeries(colStart string, rowStart int, colEnd string, rowEnd int) {
	c.Values = append(c.Values, ChartBuilderSeries{
		ColumnStart: colStart,
		RowStart:    rowStart,
		ColumnEnd:   colEnd,
		RowEnd:      rowEnd,
	})
}

func (c *ChartBuilder) GetValueSeries(i int) (string, error) {
	if i >= len(c.Values) {
		return "", fmt.Errorf("%d out of bounds", i)
	}
	v := c.Values[i]
	return fmt.Sprintf("'%s'!$%s$%d:$%s$%d",
		c.WorksheetName, v.ColumnStart, v.RowStart, v.ColumnEnd, v.RowEnd), nil
}

func (c *ChartBuilder) AddCategorySeries(colStart string, rowStart int, colEnd string, rowEnd int) {
	c.Categories = append(c.Categories, ChartBuilderSeries{
		ColumnStart: colStart,
		RowStart:    rowStart,
		ColumnEnd:   colEnd,
		RowEnd:      rowEnd,
	})
}

func (c *ChartBuilder) GetCategorySeries(i int) (string, error) {
	if i >= len(c.Categories) {
		return "", fmt.Errorf("%d out of bounds", i)
	}
	v := c.Categories[i]
	return fmt.Sprintf("'%s'!$%s$%d:$%s$%d",
		c.WorksheetName, v.ColumnStart, v.RowStart, v.ColumnEnd, v.RowEnd), nil
}

func (c *ChartBuilder) BuildChart(w *WorkSheet, cell string) error {
	//
	if len(c.Values) <= 0 {
		return fmt.Errorf("no values")
	}
	if len(c.Categories) <= 0 {
		return fmt.Errorf("no categories")
	}

	var myChartSeries []excelize.ChartSeries
	for i := 0; i < len(c.Values); i++ {
		myValues, err := c.GetValueSeries(i)
		if err != nil {
			return err
		}
		myCategories, err := c.GetCategorySeries(i)
		if err != nil {
			return err
		}
		myChartSeries = append(myChartSeries, excelize.ChartSeries{
			Values:     myValues,
			Categories: myCategories,
		})
	}

	if len(myChartSeries) <= 0 {
		return fmt.Errorf("missing chart series")
	}

	var xFont excelize.Font
	var yFont excelize.Font
	if c.xFont == nil {
		xFont = excelize.Font{
			Bold:  true,
			Color: "000000",
		}
	} else {
		xFont.Bold = c.xFont.Bold
		xFont.Color = c.xFont.Color
	}

	if c.yFont == nil {
		yFont = excelize.Font{
			Bold:  true,
			Color: "000000",
		}
	} else {
		yFont.Bold = c.yFont.Bold
		yFont.Color = c.yFont.Color
	}

	xAxis := excelize.ChartAxis{
		ReverseOrder: c.xReverse,
		Font:         xFont,
	}

	logrus.Info("xAxis:", xAxis.ReverseOrder)

	yAxis := excelize.ChartAxis{
		ReverseOrder: c.yReverse,
		Font:         xFont,
	}

	myChart := excelize.Chart{
		Type: c.Type,
		Title: excelize.ChartTitle{
			Name: c.Title,
		},
		XAxis: xAxis,
		YAxis: yAxis,
		Dimension: excelize.ChartDimension{
			Width:  c.Width,
			Height: c.Height,
		},
		Series: myChartSeries,
		PlotArea: excelize.ChartPlotArea{
			ShowCatName:     false,
			ShowLeaderLines: false,
			ShowPercent:     false,
			ShowSerName:     false,
			ShowVal:         false,
		},
		Legend: excelize.ChartLegend{
			ShowLegendKey: false,
		},
		VaryColors: &c.VaryColors,
	}

	if err := w.File.AddChart(c.WorksheetName, cell, &myChart); err != nil {
		logrus.Error(err.Error())
		return err
	}
	return nil
}
