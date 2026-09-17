package app

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenProfiles screen = iota
	screenCreate
	screenMenu
	screenPractice
	screenDrills
	screenTyping
	screenSummary
	screenHall
	screenCheat
)

type menuItem struct {
	title, desc, id string
	locked          bool
}

func (i menuItem) Title() string {
	if i.locked {
		return i.title + "  (locked)"
	}
	return i.title
}
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

// Model is the root Bubble Tea model.
type Model struct {
	styles   Styles
	store    *Store
	passages []Passage
	rng      *rand.Rand

	screen screen
	width  int
	height int

	list       list.Model
	nameInput  textinput.Model
	cheatInput textinput.Model
	xpBar      progress.Model

	active *Profile

	session       *Session
	sessionKey    string
	sessionTitle  string
	sessionHint   string
	showKeyboard  bool
	finishedEarly bool
	lastXP        int
	lastLevelUp   bool
	oldLevel      int
	praise        string
	unlockNote    string
	status        string
	createErr     string

	sessionKind   sessionKind
	journeyStepID string
	lastLitBonus  bool
	lastDrill     Drill
	boss          *BossFight
	lastBossWin   bool
	lastBossFail  bool
	bossReason    string
	cheat         []string
	cheatErr      string
}

type sessionKind int

const (
	kindDrill sessionKind = iota
	kindPlay
	kindLiterature
	kindBoss
)

// New constructs the app model. Call Init after.
func New(store *Store, passages []Passage, styles Styles) Model {
	ti := textinput.New()
	ti.Placeholder = "Your name"
	ti.CharLimit = 16
	ti.Width = 24
	ti.Prompt = "> "
	ti.PromptStyle = styles.Accent
	ti.TextStyle = styles.Value
	ti.PlaceholderStyle = styles.Subtle

	ci := textinput.New()
	ci.Placeholder = "1–100"
	ci.CharLimit = 3
	ci.Width = 8
	ci.Prompt = "> "
	ci.PromptStyle = styles.Accent
	ci.TextStyle = styles.Value

	bar := progress.New(progress.WithSolidFill(styles.Palette.Correct), progress.WithWidth(32))
	bar.ShowPercentage = false

	m := Model{
		styles:     styles,
		store:      store,
		passages:   passages,
		rng:        rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
		nameInput:  ti,
		cheatInput: ci,
		xpBar:      bar,
		width:      80,
		height:     24,
	}
	if store != nil && store.LoadErr != nil {
		m.status = store.LoadErr.Error()
	}
	m.list = newMenuList(styles, nil, 60, 12)
	if store != nil && len(store.Profiles()) == 0 {
		m.screen = screenCreate
		m.nameInput.Focus()
	} else {
		m.screen = screenProfiles
		m.setProfileItems()
	}
	return m
}

func newMenuList(st Styles, items []list.Item, w, h int) list.Model {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = st.MenuSelected
	d.Styles.SelectedDesc = st.MenuSelectedDesc
	d.Styles.NormalTitle = st.MenuNormal
	d.Styles.NormalDesc = st.MenuNormalDesc
	d.Styles.DimmedTitle = st.MenuNormal
	d.Styles.DimmedDesc = st.MenuNormalDesc
	l := list.New(items, d, w, h)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.DisableQuitKeybindings()
	return l
}

func (m Model) Init() tea.Cmd {
	if m.screen == screenCreate {
		return textinput.Blink
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resize()
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.screen != screenTyping && m.screen != screenCreate && m.screen != screenCheat {
			var hit bool
			m.cheat, hit = feedCheat(m.cheat, msg.String())
			if hit {
				return m.openCheatPrompt()
			}
		} else if m.screen != screenCheat {
			m.cheat = nil
		}
	case tickMsg:
		if m.screen == screenTyping && m.session != nil {
			if m.boss != nil && m.session.Started() && m.boss.TimedOut(m.session.Elapsed()) {
				return m.finishBoss(), nil
			}
			if !m.session.Done() {
				return m, tickCmd()
			}
		}
		return m, nil
	}

	switch m.screen {
	case screenProfiles:
		return m.updateProfiles(msg)
	case screenCreate:
		return m.updateCreate(msg)
	case screenMenu:
		return m.updateMenu(msg)
	case screenPractice:
		return m.updatePractice(msg)
	case screenDrills:
		return m.updateDrills(msg)
	case screenTyping:
		return m.updateTyping(msg)
	case screenSummary:
		return m.updateSummary(msg)
	case screenHall:
		return m.updateHall(msg)
	case screenCheat:
		return m.updateCheat(msg)
	}
	return m, nil
}

func (m *Model) resize() {
	w := m.width - 6
	h := m.height - 12
	if w < 20 {
		w = 20
	}
	if h < 5 {
		h = 5
	}
	m.list.SetSize(w, h)
}

func (m Model) updateProfiles(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			if m.active != nil {
				m.screen = screenMenu
				m.setMenuItems()
				return m, nil
			}
			return m, tea.Quit
		case "enter":
			item, ok := m.list.SelectedItem().(menuItem)
			if !ok {
				return m, nil
			}
			if item.id == "new" {
				m.screen = screenCreate
				m.createErr = ""
				m.nameInput.SetValue("")
				m.nameInput.Focus()
				return m, textinput.Blink
			}
			if err := m.store.Select(item.id); err != nil {
				m.status = err.Error()
				return m, nil
			}
			m.active = m.store.Profile(item.id)
			_ = m.store.Save()
			m.screen = screenMenu
			m.setMenuItems()
			m.status = ""
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			if len(m.store.Profiles()) > 0 {
				m.screen = screenProfiles
				m.setProfileItems()
				return m, nil
			}
			return m, tea.Quit
		case "enter":
			p, err := m.store.Create(m.nameInput.Value())
			if err != nil {
				m.createErr = err.Error()
				return m, nil
			}
			if err := m.store.Save(); err != nil {
				m.createErr = err.Error()
				return m, nil
			}
			m.active = p
			m.screen = screenMenu
			m.setMenuItems()
			m.createErr = ""
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m Model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.screen = screenProfiles
			m.setProfileItems()
			return m, nil
		case "enter":
			item, ok := m.list.SelectedItem().(menuItem)
			if !ok {
				return m, nil
			}
			return m.chooseMenu(item)
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) chooseMenu(item menuItem) (tea.Model, tea.Cmd) {
	if m.active == nil {
		m.status = "Pick a profile first."
		return m, nil
	}
	switch item.id {
	case "quit":
		return m, tea.Quit
	case "switch":
		m.screen = screenProfiles
		m.setProfileItems()
		return m, nil
	case "hall":
		m.screen = screenHall
		return m, nil
	case "play":
		return m.startPlay()
	case "drills":
		m.screen = screenDrills
		m.setDrillItems()
		return m, nil
	case "practice":
		if MaxUnlockedTier(m.active.Level) < 1 {
			m.status = "Literature unlocks at level 2. Keep going in Play!"
			return m, nil
		}
		m.screen = screenPractice
		m.setPracticeItems()
		return m, nil
	}
	return m, nil
}

func (m Model) updatePractice(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.screen = screenMenu
			m.setMenuItems()
			return m, nil
		case "enter":
			item, ok := m.list.SelectedItem().(menuItem)
			if !ok {
				return m, nil
			}
			if item.locked {
				m.status = item.desc
				return m, nil
			}
			for _, p := range m.passages {
				if p.ID == item.id {
					return m.startPassage(p, kindLiterature)
				}
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateDrills(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.screen = screenMenu
			m.setMenuItems()
			return m, nil
		case "enter":
			item, ok := m.list.SelectedItem().(menuItem)
			if !ok {
				return m, nil
			}
			for _, d := range Curriculum {
				if d.ID == item.id {
					return m.startDrill(d)
				}
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateTyping(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "esc":
		if m.sessionKind == kindBoss {
			return m.finishBoss(), nil
		}
		return m.finishSession(true), nil
	case "backspace":
		m.session.Backspace()
		return m, nil
	}
	if r, ok := typingRune(key); ok {
		pos := 0
		if m.session != nil {
			pos = m.session.Pos
		}
		m.session.Handle(r)
		if m.boss != nil {
			hit := pos < len(m.session.Missed) && m.session.Pos > pos && !m.session.Missed[pos]
			m.boss.OnKey(hit, m.session.Stats().GrossWPM)
			if m.session.Done() || (m.session.Started() && m.boss.TimedOut(m.session.Elapsed())) {
				return m.finishBoss(), nil
			}
			if m.session.Attempts() == 1 {
				return m, tickCmd()
			}
			return m, nil
		}
		if m.session.Done() {
			return m.finishSession(false), nil
		}
		if m.session.Attempts() == 1 {
			return m, tickCmd()
		}
	}
	return m, nil
}

func (m Model) updateSummary(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			return m.continueTrack()
		case "esc":
			m.screen = screenMenu
			m.setMenuItems()
			return m, nil
		}
	}
	return m, nil
}

func (m Model) continueTrack() (tea.Model, tea.Cmd) {
	switch m.sessionKind {
	case kindPlay, kindBoss:
		return m.startPlay()
	case kindLiterature:
		if m.active != nil {
			if pas, ok := Recommend(LiteraturePassages(m.passages), *m.active); ok {
				return m.startPassage(pas, kindLiterature)
			}
			m.screen = screenPractice
			m.setPracticeItems()
			return m, nil
		}
	case kindDrill:
		if m.lastDrill.ID != "" {
			return m.startDrill(m.lastDrill)
		}
	}
	m.screen = screenMenu
	m.setMenuItems()
	return m, nil
}

func (m Model) updateHall(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc", "enter":
			m.screen = screenMenu
			m.setMenuItems()
			return m, nil
		}
	}
	return m, nil
}

func (m Model) openCheatPrompt() (tea.Model, tea.Cmd) {
	if m.active == nil || m.store == nil {
		m.status = "cheat · pick a profile first"
		return m, nil
	}
	m.cheat = nil
	m.cheatErr = ""
	m.cheatInput.SetValue("")
	m.cheatInput.Focus()
	m.screen = screenCheat
	return m, textinput.Blink
}

func (m Model) updateCheat(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.cheatInput.Blur()
			m.cheatErr = ""
			m.screen = screenMenu
			m.setMenuItems()
			return m, nil
		case "enter":
			return m.applyCheatLevel(m.cheatInput.Value()), nil
		}
	}
	var cmd tea.Cmd
	m.cheatInput, cmd = m.cheatInput.Update(msg)
	return m, cmd
}

func (m Model) applyCheatLevel(raw string) Model {
	to, err := parseCheatLevel(raw)
	if err != nil {
		m.cheatErr = err.Error()
		return m
	}
	from := 0
	if m.active != nil {
		from = m.active.Level
	}
	m.active = m.store.SkipToLevel(m.active.ID, to)
	_ = m.store.Save()
	if m.active == nil {
		m.cheatErr = "could not skip"
		return m
	}
	m.cheatInput.Blur()
	m.screen = screenMenu
	m.setMenuItems()
	m.status = fmt.Sprintf("cheat · level %d → %d", from, m.active.Level)
	return m
}

func typingRune(msg tea.KeyMsg) (rune, bool) {
	if msg.Paste || msg.Alt {
		return 0, false
	}
	switch msg.String() {
	case "enter", "esc", "backspace", "tab", "up", "down", "left", "right",
		"ctrl+c", "ctrl+h":
		return 0, false
	}
	if len(msg.Runes) == 1 {
		r := msg.Runes[0]
		if unicode.IsPrint(r) {
			return r, true
		}
	}
	if msg.String() == " " {
		return ' ', true
	}
	s := msg.String()
	if len([]rune(s)) == 1 {
		r := []rune(s)[0]
		if unicode.IsPrint(r) {
			return r, true
		}
	}
	return 0, false
}

func (m Model) startPassage(p Passage, kind sessionKind) (tea.Model, tea.Cmd) {
	m.session = NewSession(p.Text)
	m.sessionKey = p.ScoreKey()
	m.sessionTitle = p.Title
	m.sessionHint = p.Source
	m.sessionKind = kind
	if kind != kindPlay {
		m.journeyStepID = ""
	}
	m.showKeyboard = m.active != nil && m.active.Level <= 3
	m.boss = nil
	m.screen = screenTyping
	m.status = ""
	return m, nil
}

func (m Model) startDrill(d Drill) (tea.Model, tea.Cmd) {
	text := Generate(d, m.rng, drillTargetLen)
	m.session = NewSession(text)
	m.sessionKey = "drill:" + d.ID
	m.sessionTitle = d.Title
	m.sessionHint = d.Desc
	m.sessionKind = kindDrill
	m.lastDrill = d
	m.boss = nil
	m.journeyStepID = ""
	m.showKeyboard = true
	m.screen = screenTyping
	m.status = ""
	return m, nil
}

func (m Model) startPlay() (tea.Model, tea.Cmd) {
	if m.active == nil {
		return m, nil
	}
	if IsBossLevel(m.active.Level) && !m.active.HasBeatBoss(m.active.Level) {
		return m.startBoss()
	}
	m.boss = nil
	step, done, _ := NextJourneyStep(*m.active)
	m.sessionKind = kindPlay
	m.journeyStepID = step.ID
	progress := fmt.Sprintf("%s  (%d/%d)", step.Title, done+1, step.Need)
	if step.Need <= 1 {
		progress = step.Title
	}
	switch step.Kind {
	case JourneyLetters:
		text := Generate(step.Drill, m.rng, drillTargetLen)
		m.session = NewSession(text)
		m.sessionKey = "play:" + step.ID
		m.sessionTitle = progress
		m.sessionHint = step.Desc
		m.showKeyboard = true
		m.screen = screenTyping
		m.status = ""
		return m, nil
	case JourneyNonsense:
		text := GenerateNonsense(KeysForLevel(m.active.Level), m.rng, drillTargetLen)
		m.session = NewSession(text)
		m.sessionKey = "play:" + step.ID
		m.sessionTitle = progress
		m.sessionHint = step.Desc
		m.showKeyboard = true
		m.screen = screenTyping
		m.status = ""
		return m, nil
	case JourneySentence:
		pas, ok := NextSentence(m.passages, *m.active)
		if !ok {
			text := GenerateNonsense(KeysForLevel(m.active.Level), m.rng, drillTargetLen)
			m.session = NewSession(text)
			m.sessionKey = "play:" + step.ID
			m.sessionTitle = "Nonsense words"
			m.sessionHint = "No sentences unlocked yet — more fake words"
			m.showKeyboard = true
			m.screen = screenTyping
			m.status = ""
			return m, nil
		}
		m.sessionTitle = progress
		um, cmd := m.startPassage(pas, kindPlay)
		next := um.(Model)
		next.sessionTitle = progress
		next.sessionHint = step.Desc + "  ·  " + pas.Title
		next.journeyStepID = step.ID
		next.sessionKind = kindPlay
		return next, cmd
	}
	return m, nil
}

func (m Model) startBoss() (tea.Model, tea.Cmd) {
	spec := BossFor(m.active.Level)
	text := BuildBossText(m.active.Level, m.passages, m.rng)
	m.boss = NewBossFight(spec, text)
	m.session = NewSession(text)
	m.sessionKey = fmt.Sprintf("boss:%d", spec.Level)
	m.sessionTitle = spec.Title
	m.sessionHint = spec.Source
	m.sessionKind = kindBoss
	m.journeyStepID = ""
	m.showKeyboard = false
	m.lastBossWin = false
	m.lastBossFail = false
	m.screen = screenTyping
	m.status = ""
	return m, nil
}

func (m Model) finishBoss() Model {
	if m.session == nil || m.active == nil || m.boss == nil {
		m.screen = screenMenu
		m.setMenuItems()
		return m
	}
	if m.session.Attempts() == 0 {
		m.screen = screenMenu
		m.setMenuItems()
		return m
	}
	st := m.session.Stats()
	finished := m.session.Done()
	win, reason := m.boss.Evaluate(st, finished, m.session.Elapsed())
	m.lastBossWin = win
	m.lastBossFail = !win
	m.bossReason = reason
	m.finishedEarly = !finished
	m.lastLitBonus = false
	m.lastLevelUp = false
	m.unlockNote = ""
	if win {
		m.praise = m.boss.WinMessage()
		xp := ComputeXP(st, true, true)
		score := Score{WPM: st.NetWPM, Accuracy: st.Accuracy, XP: xp}
		m.oldLevel = m.active.Level
		m.active = m.store.Award(m.active.ID, m.sessionKey, xp, score, m.session.KeyStats(), m.session.BigramStats())
		m.active = m.store.MarkBoss(m.active.ID, m.boss.Spec.Level)
		_ = m.store.Save()
		m.lastXP = xp
		m.lastLevelUp = m.active != nil && m.active.Level > m.oldLevel
		if m.lastLevelUp && m.active != nil {
			m.unlockNote = unlockMessage(m.oldLevel, m.active.Level)
		}
	} else {
		m.praise = m.boss.FailMessage(reason)
		m.lastXP = 0
	}
	m.screen = screenSummary
	return m
}

func (m Model) finishSession(early bool) Model {
	if m.session == nil || m.active == nil {
		m.screen = screenMenu
		m.setMenuItems()
		return m
	}
	if m.session.Attempts() == 0 {
		m.screen = screenMenu
		m.setMenuItems()
		return m
	}
	finished := m.session.Done() && !early
	st := m.session.Stats()
	litBonus := m.sessionKind == kindLiterature
	xp := ComputeXP(st, finished, litBonus)
	score := Score{WPM: st.NetWPM, Accuracy: st.Accuracy, XP: xp}
	m.oldLevel = m.active.Level
	m.active = m.store.Award(m.active.ID, m.sessionKey, xp, score, m.session.KeyStats(), m.session.BigramStats())
	m.lastLitBonus = litBonus && finished && st.Accuracy >= JourneyPassAccuracy
	if finished && st.Accuracy >= JourneyPassAccuracy && m.journeyStepID != "" {
		m.active = m.store.MarkJourney(m.active.ID, m.journeyStepID)
	}
	_ = m.store.Save()
	m.lastXP = xp
	m.lastLevelUp = m.active != nil && m.active.Level > m.oldLevel
	m.finishedEarly = early && !m.session.Done()
	m.praise = Praise(st.Accuracy, finished)
	m.unlockNote = ""
	if m.lastLevelUp && m.active != nil {
		m.unlockNote = unlockMessage(m.oldLevel, m.active.Level)
		if IsBossLevel(m.active.Level) && !m.active.HasBeatBoss(m.active.Level) {
			m.unlockNote = "A boss awaits in Play: " + BossFor(m.active.Level).Title
		}
	}
	m.lastBossWin = false
	m.lastBossFail = false
	m.boss = nil
	m.screen = screenSummary
	return m
}

func (m *Model) setProfileItems() {
	var items []list.Item
	for _, p := range m.store.Profiles() {
		items = append(items, menuItem{
			title: p.Name,
			desc:  "Level " + strconv.Itoa(p.Level) + "  ·  " + strconv.Itoa(p.TotalXP) + " XP  ·  " + strconv.Itoa(p.TotalTestsCompleted) + " tests",
			id:    p.ID,
		})
	}
	items = append(items, menuItem{title: "New typer", desc: "Create a new profile", id: "new"})
	m.list.SetItems(items)
	if last := m.store.LastProfile(); last != nil {
		for i, p := range m.store.Profiles() {
			if p.ID == last.ID {
				m.list.Select(i)
				break
			}
		}
	}
	m.resize()
}

func (m *Model) setMenuItems() {
	level := 1
	playDesc := "Your typing journey"
	if m.active != nil {
		level = m.active.Level
		playDesc = PlayDesc(*m.active)
	}
	items := []list.Item{
		menuItem{title: "Play", desc: playDesc, id: "play"},
		menuItem{title: "Drills", desc: "Home-row practice and finger placement", id: "drills"},
		menuItem{title: "Literature", desc: literatureDesc(level), id: "practice", locked: MaxUnlockedTier(level) < 1},
		menuItem{title: "Hall of Fame", desc: "Compare high scores side by side", id: "hall"},
		menuItem{title: "Switch profile", desc: "Play as someone else", id: "switch"},
		menuItem{title: "Quit", desc: "See you next time", id: "quit"},
	}
	m.list.SetItems(items)
	m.status = ""
	m.resize()
}

func (m *Model) setPracticeItems() {
	level := 1
	if m.active != nil {
		level = m.active.Level
	}
	max := MaxUnlockedTier(level)
	var items []list.Item
	for _, p := range LiteraturePassages(m.passages) {
		locked := p.Tier > max
		desc := p.Category + "  ·  " + p.Source
		if locked {
			desc = "Unlocks at level " + strconv.Itoa(tierUnlockLevel(p.Tier))
		} else {
			desc += "  ·  bonus XP"
		}
		items = append(items, menuItem{
			title:  p.Title,
			desc:   desc,
			id:     p.ID,
			locked: locked,
		})
	}
	m.list.SetItems(items)
	m.resize()
}

func (m *Model) setDrillItems() {
	var items []list.Item
	level := 1
	if m.active != nil {
		level = m.active.Level
	}
	for _, d := range LetterDrills() {
		if level < d.MinLevel {
			continue
		}
		items = append(items, menuItem{title: d.Title, desc: d.Desc, id: d.ID})
	}
	m.list.SetItems(items)
	m.resize()
}

func literatureDesc(level int) string {
	if MaxUnlockedTier(level) < 1 {
		return "Unlocks at level 2 — extra XP for classic passages"
	}
	return "Classic passages. Finish well for bonus XP."
}

func tierUnlockLevel(tier int) int {
	switch tier {
	case 1:
		return 2
	case 2:
		return 3
	default:
		return 5
	}
}

func unlockMessage(oldLevel, newLevel int) string {
	var bits []string
	for lvl := oldLevel + 1; lvl <= newLevel; lvl++ {
		switch lvl {
		case 2:
			bits = append(bits, "Short sentences join Play, and Literature mode is open!")
		case 3:
			bits = append(bits, "Longer literature and scripture unlocked!")
		case 5:
			bits = append(bits, "Poetry and long passages unlocked!")
		}
	}
	if len(bits) == 0 {
		return "You leveled up!"
	}
	return strings.Join(bits, " ")
}

// Praise is encouraging copy for the summary screen.
func Praise(acc float64, finished bool) string {
	if !finished {
		return "Nice try! Every key you type makes you stronger."
	}
	switch {
	case acc >= 1:
		return "PERFECT! You nailed every letter!"
	case acc >= 0.95:
		return "Awesome accuracy!"
	case acc >= 0.85:
		return "Great job!"
	case acc >= 0.7:
		return "You're getting it — keep going!"
	default:
		return "Practice makes you stronger. Try that one again!"
	}
}
