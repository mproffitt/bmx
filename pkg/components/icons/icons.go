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

package icons

const (
	Down  string = "↓"
	Enter string = "↩"
	Left  string = "←"
	Right string = "→"
	Shift string = "⇧"
	Space string = "␣"
	Tab   string = "↹"
	Up    string = "↑"
)

// Compound icons
const ShiftTab = Shift + " " + Tab

var (
	DigitalNumbers = [10]rune{
		'🯰', '🯱', '🯲', '🯳', '🯴',
		'🯵', '🯶', '🯷', '🯸', '🯹',
	}

	HsquareNumbers = [10]rune{
		'󰎣', '󰎦', '󰎩', '󰎬', '󰎮',
		'󰎰', '󰎵', '󰎸', '󰎻', '󰎾',
	}
)

const (
	ActivityIcon       = '󱅫'
	ActiveTerminalIcon = ''
	ApplicationIcon    = ''
	BellIcon           = '󰂞'
	CurrentIcon        = '󰖯'
	Ellipis            = '…'
	Error              = 'ⓧ'
	GitIcon            = '󰊢'
	HostIcon           = '󰒋'
	Info               = 'ⓘ'
	LastIcon           = '󰖰'
	MarkedIcon         = '󰃀'
	SilenceIcon        = '󰂛'
	TerminalIcon       = ''
	Tick               = '✓'
	UserIcon           = ''
	Warning            = '⚠'
	ZoomIcon           = '󰁌'
)

const (
	Ellipsis   = '…'
	Kubernetes = '󱃾'
)
