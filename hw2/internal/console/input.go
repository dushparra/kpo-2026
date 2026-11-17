package console

import (
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
)

func (c *terminal) readLine(prompt string) (string, error) {
	c.write("%s", prompt)
	if c.ioError != nil {
		return "", c.ioError
	}
	if !c.scanner.Scan() {
		c.ioError = c.scanner.Err()
		if c.ioError == nil {
			c.ioError = io.EOF
			c.inputClosed = true
		}
		return "", c.ioError
	}
	return strings.TrimSpace(c.scanner.Text()), nil
}

func (c *terminal) readNumber(prompt string) (float64, error) {
	line, err := c.readLine(prompt)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseFloat(strings.ReplaceAll(line, ",", "."), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, errors.New("нужно ввести конечное число, например 7,5 или 7.5")
	}
	return value, nil
}
