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
