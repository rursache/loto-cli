package client

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/rursache/loto-cli/models"
)

// resultsSectionClass identifies the results row on the loto.ro homepage.
const resultsSectionClass = "vc_custom_1643109784313"

// GetResults scrapes the latest lottery extraction results from the loto.ro main page.
//
// The page structure has 3 columns (vc_col-sm-4) inside the results row:
//
//	Column 1: Loto 6/49 + Noroc     (logo: Loto_6_49__noroc.png)
//	Column 2: Loto 5/40 + Super Noroc (logo: Loto_5_40__super_noroc.png)
//	Column 3: Joker                  (logo: joker__noroc_plus.png)
//
// Each column contains Ninja Tables (footable_*):
//   - First visible 6-col table: main draw numbers
//   - 1-col table with a date (DD-MM-YYYY): draw date
//   - 1-col table with space-separated digits: Noroc/Super Noroc number
//   - Hidden tables (parent has "ascuns" class): second draw (ignored)
func (c *Client) GetResults() ([]models.Extraction, error) {
	req, err := c.newRequest("GET", baseLotoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for results: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch results page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d when fetching results", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse results HTML: %w", err)
	}

	resultsRow := doc.Find("div[class*='" + resultsSectionClass + "']").First()
	if resultsRow.Length() == 0 {
		return nil, fmt.Errorf("results section not found on page")
	}

	var extractions []models.Extraction

	// Iterate over the 3 columns
	resultsRow.Find("div.vc_col-sm-4").Each(func(colIdx int, col *goquery.Selection) {
		game := identifyGameFromColumn(col)
		if game == "" {
			return
		}

		var numbers []int
		var date string
		var norocNumber string

		// Find all visible tables in this column (skip hidden "ascuns" ones)
		col.Find("table").Each(func(_ int, table *goquery.Selection) {
			if isHidden(table) {
				return
			}

			cells := extractCells(table)
			if len(cells) == 0 {
				return
			}

			// Single-cell table: either a date or a Noroc number
			if len(cells) == 1 {
				val := strings.TrimSpace(cells[0])
				if isDate(val) {
					if date == "" {
						date = val
					}
				} else if isSpacedDigits(val) {
					norocNumber = strings.ReplaceAll(val, " ", "")
				}
				return
			}

			// Multi-cell table: draw numbers (take only the first visible one)
			if len(numbers) == 0 {
				for _, cell := range cells {
					if n, err := strconv.Atoi(strings.TrimSpace(cell)); err == nil {
						numbers = append(numbers, n)
					}
				}
			}
		})

		// Build the main game extraction
		if len(numbers) > 0 {
			ext := models.Extraction{
				Game:    game,
				Date:    date,
				Numbers: numbers,
			}

			// For Joker: last number in the 6-cell row is the Joker bonus
			if game == models.GameJoker && len(numbers) == 6 {
				ext.Numbers = numbers[:5]
				ext.Bonus = numbers[5:]
			}

			extractions = append(extractions, ext)
		}

		// Build the Noroc/Super Noroc extraction from the spaced-digit table
		if norocNumber != "" {
			norocGame := norocGameForColumn(game)
			if norocGame != "" {
				// Convert each digit to an int for the Numbers field
				var digits []int
				for _, ch := range norocNumber {
					if d, err := strconv.Atoi(string(ch)); err == nil {
						digits = append(digits, d)
					}
				}
				extractions = append(extractions, models.Extraction{
					Game:    norocGame,
					Date:    date,
					Numbers: digits,
				})
			}
		}
	})

	return extractions, nil
}

// identifyGameFromColumn determines the game type from the logo image in the column.
func identifyGameFromColumn(col *goquery.Selection) models.Game {
	var game models.Game
	col.Find("img").Each(func(_ int, img *goquery.Selection) {
		if game != "" {
			return
		}
		src, _ := img.Attr("data-src")
		if src == "" {
			src, _ = img.Attr("src")
		}
		lower := strings.ToLower(src)
		switch {
		case strings.Contains(lower, "loto_6_49") || strings.Contains(lower, "logo649"):
			game = models.GameLoto649
		case strings.Contains(lower, "loto_5_40") || strings.Contains(lower, "logo540"):
			game = models.GameLoto540
		case strings.Contains(lower, "joker"):
			game = models.GameJoker
		}
	})
	return game
}

// norocGameForColumn maps a main game to its associated Noroc game.
func norocGameForColumn(mainGame models.Game) models.Game {
	switch mainGame {
	case models.GameLoto649:
		return models.GameNoroc
	case models.GameLoto540:
		return models.GameSuperNoroc
	default:
		return ""
	}
}

// isHidden checks if a table (or its ancestor) has the "ascuns" class, meaning it's hidden.
func isHidden(table *goquery.Selection) bool {
	hidden := false
	table.Parents().Each(func(_ int, parent *goquery.Selection) {
		class, _ := parent.Attr("class")
		if strings.Contains(class, "ascuns") {
			hidden = true
		}
	})
	return hidden
}

// extractCells returns all td text values from a table's tbody.
func extractCells(table *goquery.Selection) []string {
	var cells []string
	table.Find("tbody td").Each(func(_ int, td *goquery.Selection) {
		cells = append(cells, strings.TrimSpace(td.Text()))
	})
	return cells
}

// isDate checks if a string matches DD-MM-YYYY or DD.MM.YYYY format.
func isDate(s string) bool {
	if len(s) != 10 {
		return false
	}
	for i, ch := range s {
		if i == 2 || i == 5 {
			if ch != '-' && ch != '.' {
				return false
			}
		} else {
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}
	return true
}

// isSpacedDigits checks if a string is space-separated single digits (e.g. "5 3 8 6 5 3 5").
func isSpacedDigits(s string) bool {
	if len(s) < 3 {
		return false
	}
	hasSpace := false
	for _, ch := range s {
		if ch == ' ' {
			hasSpace = true
		} else if ch < '0' || ch > '9' {
			return false
		}
	}
	return hasSpace
}
