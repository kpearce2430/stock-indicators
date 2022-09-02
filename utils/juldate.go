package utils

import (
	"fmt"
	"time"
)

func JulDate() string {

	t := time.Now()
	return fmt.Sprintf("%d%00d", t.Year(), t.YearDay())
}
