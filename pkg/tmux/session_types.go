package tmux

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Session struct {
	Name     string
	Windows  int
	Created  time.Time
	Attached bool
}

func (s Session) Title() string {
	return s.Name
}

func (s Session) Description() string {
	date := s.Created.Format(time.ANSIC)
	if s.Attached {
		return fmt.Sprintf("active\n%s", date)
	}
	return date
}
func (s Session) FilterValue() string { return s.Name }

func NewSessionFromString(session string) Session {
	re := regexp.MustCompile(`(?P<name>.*):\s(?P<windows>\d)\swindows\s+\(created\s(?P<created>[^)]*)\)\s?(?P<attached>.*)`)
	matches := re.FindStringSubmatch(session)

	params := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i > 0 && i < len(matches) {
			params[name] = matches[i]
		}
	}

	var (
		details                          = Session{}
		name, windows, created, attached string
		ok                               bool
		err                              error
	)
	if name, ok = params["name"]; ok {
		details.Name = name
	}
	if windows, ok = params["windows"]; ok {
		var count int
		count, err = strconv.Atoi(windows)
		if err != nil {
			count = 0
		}
		details.Windows = count
	}

	if created, ok = params["created"]; ok {
		date, err := time.Parse(time.ANSIC, created)
		if err != nil {
			fmt.Println(err)
			date = time.Now()
		}
		details.Created = date
	}

	if attached, ok = params["attached"]; ok && attached != "" {
		details.Attached = true
	}
	return details
}
