package models

// Game represents a lottery game type
type Game string

const (
	GameLoto649   Game = "Loto 6/49"
	GameLoto540   Game = "Loto 5/40"
	GameJoker     Game = "Joker"
	GameNoroc     Game = "Noroc"
	GameSuperNoroc Game = "Super Noroc"
)

// Extraction represents a single lottery draw result
type Extraction struct {
	Game    Game
	Date    string
	Numbers []int
	Bonus   []int // Joker bonus number, Noroc/Super Noroc number
}

// Ticket represents a purchased lottery ticket
type Ticket struct {
	OrderID   string
	TicketID  string
	Game      Game
	Price     string // e.g. "24,50 RON"
	DrawDate  string // e.g. "15.02.2026"
	Status    TicketStatus
	PlayedAt  string // e.g. "Jo 12 feb 2026, Ora 18:58"
	DetailURL string
}

// TicketStatus represents the status of a ticket
type TicketStatus int

const (
	StatusUnknown TicketStatus = iota
	StatusPending
	StatusWon
	StatusLost
)

// String returns the display string for a ticket status
func (s TicketStatus) String() string {
	switch s {
	case StatusPending:
		return "Pending"
	case StatusWon:
		return "Won"
	case StatusLost:
		return "Lost"
	default:
		return "Unknown"
	}
}

// GameFromImagePath maps an image filename to a Game type
func GameFromImagePath(path string) Game {
	switch {
	case contains(path, "logo49"):
		return GameLoto649
	case contains(path, "logo40"):
		return GameLoto540
	case contains(path, "logo45"):
		return GameJoker
	default:
		return Game("Unknown")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
