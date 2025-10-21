package pkg

const Name = "All Days"

var Days = []string{"Sunday"}

func init() {
	Days = append(Days, "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday")
}

func IsWeekend(day string) bool {
	return day == "Saturday" || day == "Sunday"
}
