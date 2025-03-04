// Copyright (c) 2025 Martin Proffitt <mprooffitt@choclab.net>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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
