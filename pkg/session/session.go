package session

import (
	"math"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/dialog"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/mproffitt/bmx/pkg/kubernetes/ui/panel"
	"github.com/mproffitt/bmx/pkg/repos"
	"github.com/mproffitt/bmx/pkg/tmux"
)

const (
	listWidth         = 26
	previewWidth      = 80
	previewHeight     = 30
	paddingMultiplier = 5

	/*sessionList             = "list"
	previewPane             = "preview"
	contextPane             = "kubernetes"
	overlay                 = "overlay"
	dialogp                 = "dialog"*/
	kubernetesSessionHeight = .4
)

type FocusType int

const (
	sessionList FocusType = iota
	previewPane
	contextPane
	overlay
	dialogp
	helpd
)

type model struct {
	config   *config.Config
	context  tea.Model
	deleting bool
	dialog   tea.Model
	filter   textinput.Model
	focused  FocusType
	height   int
	keymap   keyMap
	list     list.Model
	overlay  *overlayContainer
	preview  viewport.Model
	session  tmux.Session

	styles styles
	width  int
}

type styles struct {
	sessionlist     lipgloss.Style
	viewportNormal  lipgloss.Style
	viewportFocused lipgloss.Style
	delegates       delegates
}

type delegates struct {
	normal list.DefaultDelegate
	shaded list.DefaultDelegate
}

func New(c *config.Config) *model {
	items := []list.Item{}
	m := model{
		config:  c,
		filter:  textinput.New(),
		focused: sessionList,
		keymap:  mapKeys(),
		list:    list.New(items, list.NewDefaultDelegate(), 0, 0),
		preview: viewport.New(0, 0),
		styles: styles{
			sessionlist: lipgloss.NewStyle().Margin(1, 2).Width(listWidth),
			viewportNormal: lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(c.Style.BorderFgColor)).
				PaddingRight(2),
			viewportFocused: lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(c.Style.FocusedColor)).
				PaddingRight(2),
			delegates: delegates{},
		},
	}

	m.styles.delegates.normal = m.createListNormalDelegate()
	m.styles.delegates.shaded = m.createListShadedDelegate()

	m.setItems()
	m.list.SetDelegate(m.styles.delegates.normal)
	m.list.SetShowFilter(false)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowHelp(false)
	m.list.SetShowPagination(true)
	m.list.SetShowStatusBar(false)
	m.list.SetShowTitle(false)

	return &m
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Overlay() helpers.UseOverlay {
	repos := repos.New(m.config, repos.RepoCallback)
	width := int(math.Ceil(float64(m.width) * .8))
	height := int(math.Ceil(float64(m.height) * .75))
	repos.SetSize(width, height)
	return repos.Overlay()
}

func (m *model) GetSize() (int, int) {
	return m.width, m.height
}

func (m *model) getSessionKubeconfig(session string) string {
	kubeconfig := tmux.GetTmuxEnvVar(session, "KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = kubernetes.DefaultConfigFile()
	}
	return kubeconfig
}

func (m *model) handleDialog(status dialog.Status) (tea.Cmd, error) {
	var (
		cmd tea.Cmd
		err error
	)
	m.dialog = nil

	if m.deleting {
		switch status {
		case dialog.Confirm:
			err = kubernetes.DeleteConfig(m.session.Name)
			if err == nil {
				err = tmux.KillSession(m.session.Name)
			}
			m.list.Select(0)
			fallthrough
		case dialog.Cancel:
			m.deleting = false
		}
		m.setItems()
	}

	if m.overlay != nil {
		cmd = helpers.OverlayCmd(status)
	}
	return cmd, err
}

func (m *model) setItems() {
	sessions := tmux.ListSessions()
	items := make([]list.Item, len(sessions))
	sort.SliceStable(sessions, func(i, j int) bool {
		if sessions[i].Attached != sessions[j].Attached {
			return sessions[i].Attached
		}
		return sessions[i].Name < sessions[j].Name
	})
	for i, s := range sessions {
		items[i] = s
	}
	_ = m.list.SetItems(items)
}

// TODO: There is something in this that prevents
// the model from resizing correctly on small terminal
// windows. It would be good to understand what this is
// in order to better lay-out the frame and not have to
// disable if the term size is too small
func (m *model) resize() {
	if m.width < config.MinWidth || m.height < config.MinHeight {
		return
	}
	_, v := m.styles.sessionlist.GetFrameSize()
	m.list.SetSize(listWidth, m.height-(paddingMultiplier*v))
	w, _ := m.styles.viewportNormal.GetFrameSize()
	m.preview.Width = (m.width - listWidth) - (paddingMultiplier * w)
	m.preview.Height = (m.height - 2)
	if m.config.ManageSessionKubeContext {
		sessionHeight := int(math.Ceil(float64(m.height) * kubernetesSessionHeight))
		sessionHeight -= (panel.PanelTitle + panel.PanelFooter)
		rows, cols, colWidth := 0, 0, 0
		{
			cols = int(math.Floor(float64(m.preview.Width-2) / float64(panel.KubernetesListWidth)))
			colWidth = int(math.Floor(float64(m.preview.Width-2) / float64(cols)))
			rows = int(math.Floor(float64(sessionHeight)/float64(panel.KubernetesRowHeight))) - 1
		}

		if m.context == nil {
			session := m.list.SelectedItem().(list.DefaultItem).Title()
			m.context = panel.NewKubectxPane(m.config, session, rows, cols, colWidth)
		}
		m.preview.Height = (m.height - 3) - sessionHeight - 1
		m.context = m.context.(*panel.Model).SetSize(m.preview.Width, sessionHeight, colWidth)
	}
}
