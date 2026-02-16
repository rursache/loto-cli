package client

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/rursache/loto-cli/models"
)

const (
	ticketHistoryBaseURL = baseBileteURL + "/history/ticket"
	ticketsPerPage       = 6
)

// GetTickets fetches a single page of ticket history and returns the tickets and total count
func (c *Client) GetTickets(page int) ([]models.Ticket, int, error) {
	url := fmt.Sprintf("%s?page_no=%d", ticketHistoryBaseURL, page)

	req, err := c.newRequest("GET", url)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch tickets: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Parse total count from pagination text: "Showing 1 to 6 of 81 results"
	total := parseTotalCount(doc)

	var tickets []models.Ticket

	doc.Find("div.ticket-preview").Each(func(_ int, card *goquery.Selection) {
		ticket := parseTicketCard(card)
		tickets = append(tickets, ticket)
	})

	return tickets, total, nil
}

// GetAllTickets fetches all pages of ticket history and enriches won tickets with prize amounts
func (c *Client) GetAllTickets() ([]models.Ticket, error) {
	firstPage, total, err := c.GetTickets(1)
	if err != nil {
		return nil, err
	}

	if total == 0 || len(firstPage) == 0 {
		return firstPage, nil
	}

	allTickets := make([]models.Ticket, 0, total)
	allTickets = append(allTickets, firstPage...)

	totalPages := int(math.Ceil(float64(total) / float64(ticketsPerPage)))

	for page := 2; page <= totalPages; page++ {
		tickets, _, err := c.GetTickets(page)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
		}
		if len(tickets) == 0 {
			break
		}
		allTickets = append(allTickets, tickets...)
	}

	// Fetch prize amounts for won tickets
	for i := range allTickets {
		if allTickets[i].Status == models.StatusWon && allTickets[i].DetailURL != "" {
			prize, err := c.GetTicketPrize(allTickets[i].DetailURL)
			if err == nil && prize != "" {
				allTickets[i].Prize = prize
			}
		}
	}

	return allTickets, nil
}

// GetTicketPrize fetches a ticket detail page and extracts the total prize amount.
// The prize is in a <tfoot> row with "TOTAL CÂȘTIG" label.
func (c *Client) GetTicketPrize(detailURL string) (string, error) {
	req, err := c.newRequest("GET", detailURL)
	if err != nil {
		return "", err
	}
	req.Header.Set("Referer", ticketHistoryBaseURL+"?page_no=1")

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	var prize string
	doc.Find("tfoot tr").Each(func(_ int, row *goquery.Selection) {
		text := strings.TrimSpace(row.Text())
		if strings.Contains(text, "TOTAL") {
			// The last <th> in the row contains the prize amount
			row.Find("th").Each(func(_ int, th *goquery.Selection) {
				val := strings.TrimSpace(th.Text())
				if strings.Contains(val, "RON") {
					prize = val
				}
			})
		}
	})

	return prize, nil
}

// parseTotalCount extracts the total ticket count from the pagination text
func parseTotalCount(doc *goquery.Document) int {
	var total int

	doc.Find("p.small.text-muted").Each(func(_ int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "results") {
			re := regexp.MustCompile(`of\s+(\d+)\s+results`)
			matches := re.FindStringSubmatch(text)
			if len(matches) == 2 {
				total, _ = strconv.Atoi(matches[1])
			}
		}
	})

	return total
}

// parseTicketCard extracts a Ticket from a single ticket-preview card
func parseTicketCard(card *goquery.Selection) models.Ticket {
	var ticket models.Ticket

	items := card.Find("li.list-group-item")

	// First li: game image and price
	firstItem := items.First()
	imgSrc, _ := firstItem.Find("img").Attr("src")
	ticket.Game = models.GameFromImagePath(imgSrc)
	ticket.Price = parsePrice(firstItem.Find("span.price"))

	// Iterate over remaining list items to find fields by label text
	items.Each(func(_ int, li *goquery.Selection) {
		text := strings.TrimSpace(li.Text())

		switch {
		case strings.Contains(text, "ID Comandă"):
			ticket.OrderID = strings.TrimSpace(li.Find("span").Text())

		case strings.Contains(text, "ID Bilet"):
			ticket.TicketID = strings.TrimSpace(li.Find("span").Text())

		case strings.Contains(text, "Tragerea"):
			ticket.DrawDate = strings.TrimSpace(li.Find("span").Text())

		case strings.Contains(text, "Stare Bilet"):
			badge := strings.TrimSpace(li.Find("span.badge").Text())
			ticket.Status = parseTicketStatus(badge)
		}
	})

	// Detail URL
	detailLink := card.Find("a[href*='ticket/details']")
	if href, exists := detailLink.Attr("href"); exists {
		ticket.DetailURL = href
	}

	// Played date from card footer
	playedText := strings.TrimSpace(card.Find(".card-footer small").Text())
	playedText = strings.TrimPrefix(playedText, "Jucat ")
	ticket.PlayedAt = strings.TrimSpace(playedText)

	return ticket
}

// parsePrice extracts the price from a span.price element
// Format: "24<sup>,50</sup> <em>ron</em>" -> "24,50 RON"
func parsePrice(priceSpan *goquery.Selection) string {
	supText := strings.TrimSpace(priceSpan.Find("sup").Text())
	// Remove child elements to get only direct text content
	priceClone := priceSpan.Clone()
	priceClone.Find("sup").Remove()
	priceClone.Find("em").Remove()
	integerPart := strings.TrimSpace(priceClone.Text())

	if integerPart == "" {
		return ""
	}

	if supText != "" {
		// supText contains the comma and decimals, e.g. ",50"
		return integerPart + supText + " RON"
	}

	return integerPart + " RON"
}

// parseTicketStatus maps Romanian status text to a TicketStatus
func parseTicketStatus(status string) models.TicketStatus {
	switch strings.TrimSpace(status) {
	case "Necâștigător":
		return models.StatusLost
	case "Câștigător":
		return models.StatusWon
	default:
		return models.StatusPending
	}
}
