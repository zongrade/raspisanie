package raspisanietypes

type Class struct {
	Code        string
	Name        string
	TeacherFull string
	Teacher     string
	Form        string
}

type Group struct {
	Code string
	Name string
}

type Room struct {
	Code uint16
	Name string
}

type RaspisanieDay struct {
	Day       uint8
	DayNumber uint8
	ParTime   `json:"Time"`
	Class     `json:"Class"`
	Group     `json:"Group"`
	Room      `json:"Room"`
}

type ParTime struct {
	Time     string
	TimeFrom string
	TimeTo   string
	Code     uint8
}

type RaspisanieGroup struct {
	Times          []ParTime       `json:"Times"`
	RaspisanieDays []RaspisanieDay `json:"Data"`
	Semestr        string
}
