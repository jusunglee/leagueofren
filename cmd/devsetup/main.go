// devsetup is a comprehensive developer environment setup wizard.
// It installs dependencies (Docker, Atlas, Air), starts PostgreSQL,
// applies the database schema, configures git hooks, and collects
// environment variables. Run via `make setup`.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type step int

const (
	stepDocker step = iota
	stepAtlas
	stepAir
	stepEnv
	stepPostgres
	stepSchema
	stepGitHooks
	stepComplete
)

var stepNames = []string{
	"Docker",
	"Atlas CLI",
	"Air (hot reload)",
	"Environment (.env)",
	"PostgreSQL",
	"Database Schema",
	"Git Hooks",
	"Complete",
}

type envField int

const (
	fieldDiscordToken envField = iota
	fieldGuildID
	fieldRiotKey
	fieldLLMProvider
	fieldAnthropicKey
	fieldGoogleKey
	fieldDone
)

var envFieldNames = []string{
	"Discord Bot Token",
	"Discord Guild ID (optional)",
	"Riot API Key",
	"LLM Provider",
	"Anthropic API Key",
	"Google API Key (optional)",
}

type model struct {
	step         step
	envField     envField
	textInput    textinput.Model
	envValues    map[envField]string
	status       string
	err          error
	width        int
	height       int
	skippedSteps map[step]bool
	needsRelogin bool
}

type stepDoneMsg struct {
	skipped bool
}
type stepErrorMsg struct{ err error }
type dockerReloginMsg struct{}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	stepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	activeStepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Underline(true)

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)
)

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60

	return model{
		step:         stepDocker,
		envField:     fieldDiscordToken,
		textInput:    ti,
		envValues:    make(map[envField]string),
		skippedSteps: make(map[step]bool),
	}
}

func (m model) Init() tea.Cmd {
	return m.runCurrentStep()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			if m.step == stepEnv && m.envField < fieldDone {
				return m.handleEnvInput()
			}
			if m.step == stepComplete {
				return m, tea.Quit
			}
		case "tab":
			if m.step == stepEnv && m.envField < fieldDone {
				return m.skipEnvField()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case stepDoneMsg:
		m.skippedSteps[m.step] = msg.skipped
		m.step++
		if m.step == stepEnv && envFileExists() {
			m.skippedSteps[stepEnv] = true
			m.step++
		}
		if m.step <= stepComplete {
			return m, m.runCurrentStep()
		}

	case stepErrorMsg:
		m.err = msg.err
		return m, nil

	case dockerReloginMsg:
		m.needsRelogin = true
		return m, nil
	}

	if m.step == stepEnv && m.envField < fieldDone {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("League of Ren - Developer Setup"))
	s.WriteString("\n\n")

	s.WriteString(m.renderProgress())
	s.WriteString("\n\n")

	s.WriteString(m.renderStepContent())
	s.WriteString("\n\n")

	s.WriteString(subtleStyle.Render("enter=continue • esc/ctrl+c=quit"))
	if m.step == stepEnv && m.envField < fieldDone {
		s.WriteString(subtleStyle.Render(" • tab=skip"))
	}

	return boxStyle.Render(s.String())
}

func (m model) renderProgress() string {
	var dots []string
	for i := 0; i <= int(stepComplete); i++ {
		if i < int(m.step) {
			dots = append(dots, completedStyle.Render("●"))
		} else if i == int(m.step) {
			dots = append(dots, activeStepStyle.Render("●"))
		} else {
			dots = append(dots, stepStyle.Render("○"))
		}
	}
	progress := strings.Join(dots, " ")

	stepLabel := ""
	if m.step <= stepComplete {
		stepLabel = fmt.Sprintf("Step %d of %d: %s", m.step+1, len(stepNames), stepNames[m.step])
	}

	return fmt.Sprintf("[%s]  %s", progress, activeStepStyle.Render(stepLabel))
}

func (m model) renderStepContent() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.needsRelogin {
		return errorStyle.Render("Docker installed! Please log out and back in for group changes to take effect.\nThen run 'make setup' again.")
	}

	switch m.step {
	case stepDocker:
		return "Checking Docker installation..."
	case stepAtlas:
		return "Checking Atlas CLI installation..."
	case stepAir:
		return "Checking Air installation..."
	case stepEnv:
		return m.renderEnvStep()
	case stepPostgres:
		return "Starting PostgreSQL container..."
	case stepSchema:
		return "Applying database schema..."
	case stepGitHooks:
		return "Configuring git hooks..."
	case stepComplete:
		return m.renderComplete()
	}
	return ""
}

func (m model) renderEnvStep() string {
	if m.envField >= fieldDone {
		return completedStyle.Render("Environment configured!")
	}

	var s strings.Builder
	s.WriteString("Configure your environment:\n\n")

	fieldName := envFieldNames[m.envField]
	s.WriteString(fmt.Sprintf("%s:\n", activeStepStyle.Render(fieldName)))

	switch m.envField {
	case fieldDiscordToken:
		s.WriteString(fmt.Sprintf("  1. Go to %s\n", linkStyle.Render("https://discord.com/developers/applications")))
		s.WriteString("  2. Click \"New Application\" → name it → Create\n")
		s.WriteString("  3. Go to \"Bot\" → Click \"Reset Token\" → Copy the token\n")
		s.WriteString("  4. Under \"Privileged Gateway Intents\", enable:\n")
		s.WriteString("     • Message Content Intent\n")
		s.WriteString("  5. Go to \"OAuth2\" → \"URL Generator\"\n")
		s.WriteString("     • Scopes: bot, applications.commands\n")
		s.WriteString("     • Bot Permissions: Send Messages, Embed Links, Read Message History\n")
		s.WriteString("  6. Copy the generated URL and open it to invite the bot\n")
		s.WriteString("\n")
		s.WriteString(subtleStyle.Render("  Paste the token from step 3:\n"))
	case fieldGuildID:
		s.WriteString("  For faster command registration during development:\n")
		s.WriteString("  1. Enable Developer Mode in Discord (Settings → Advanced)\n")
		s.WriteString("  2. Right-click your server → Copy Server ID\n")
		s.WriteString("\n")
		s.WriteString(subtleStyle.Render("  Paste the Server ID (or tab to skip):\n"))
	case fieldRiotKey:
		s.WriteString(fmt.Sprintf("  1. Go to %s\n", linkStyle.Render("https://developer.riotgames.com/")))
		s.WriteString("  2. Sign in with your Riot account\n")
		s.WriteString("  3. Register a new product or use Development API Key\n")
		s.WriteString("\n")
		s.WriteString(subtleStyle.Render("  Paste your Riot API key:\n"))
	case fieldLLMProvider:
		s.WriteString("  Choose your translation provider:\n")
		s.WriteString("  • anthropic - Claude (recommended, better quality)\n")
		s.WriteString("  • google - Gemini (faster, cheaper)\n")
		s.WriteString("\n")
		s.WriteString(subtleStyle.Render("  Enter 'anthropic' or 'google':\n"))
	case fieldAnthropicKey:
		s.WriteString(fmt.Sprintf("  1. Go to %s\n", linkStyle.Render("https://console.anthropic.com/")))
		s.WriteString("  2. Sign up/in → Go to API Keys → Create Key\n")
		s.WriteString("\n")
		s.WriteString(subtleStyle.Render("  Paste your Anthropic API key:\n"))
	case fieldGoogleKey:
		s.WriteString(fmt.Sprintf("  1. Go to %s\n", linkStyle.Render("https://aistudio.google.com/apikey")))
		s.WriteString("  2. Click \"Create API Key\"\n")
		s.WriteString("\n")
		s.WriteString(subtleStyle.Render("  Paste your Google API key (or tab to skip):\n"))
	}

	s.WriteString("\n")
	s.WriteString(m.textInput.View())

	return s.String()
}

func (m model) renderComplete() string {
	var s strings.Builder
	s.WriteString(completedStyle.Render("✓ Setup complete!"))
	s.WriteString("\n\n")

	skipped := 0
	for _, v := range m.skippedSteps {
		if v {
			skipped++
		}
	}

	if skipped > 0 {
		s.WriteString(subtleStyle.Render(fmt.Sprintf("(%d steps were already configured)\n\n", skipped)))
	}

	s.WriteString("Next steps:\n")
	s.WriteString("  1. Run " + activeStepStyle.Render("make watch") + " to start the bot with hot reload\n")
	s.WriteString("  2. Invite your bot to a Discord server\n")
	s.WriteString("  3. Use /subscribe to track a player\n")

	return s.String()
}

func (m model) handleEnvInput() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.textInput.Value())

	if m.envField == fieldGuildID || m.envField == fieldGoogleKey {
		m.envValues[m.envField] = value
	} else if value == "" {
		return m, nil
	} else {
		if m.envField == fieldLLMProvider && value != "anthropic" && value != "google" {
			return m, nil
		}
		m.envValues[m.envField] = value
	}

	m.textInput.SetValue("")
	m.envField++

	if m.envField == fieldDone {
		return m, m.writeEnvFile()
	}

	return m, nil
}

func (m model) skipEnvField() (tea.Model, tea.Cmd) {
	if m.envField == fieldGuildID || m.envField == fieldGoogleKey {
		m.envValues[m.envField] = ""
		m.textInput.SetValue("")
		m.envField++

		if m.envField == fieldDone {
			return m, m.writeEnvFile()
		}
	}
	return m, nil
}

func (m model) writeEnvFile() tea.Cmd {
	return func() tea.Msg {
		provider := m.envValues[fieldLLMProvider]
		model := "claude-sonnet-4-5-20250929"
		if provider == "google" {
			model = "gemini-2.0-flash"
		}

		content := fmt.Sprintf(`# Generated by setup tool

# Database Configuration
DATABASE_URL=postgres://leagueofren:localdev123@localhost:5432/leagueofren?sslmode=disable

# Discord Configuration
DISCORD_TOKEN=%s
DISCORD_GUILD_ID=%s

# Riot API Configuration
RIOT_API_KEY=%s

# LLM API Configuration
LLM_PROVIDER=%s
LLM_MODEL=%s
ANTHROPIC_API_KEY=%s
GOOGLE_API_KEY=%s
`,
			m.envValues[fieldDiscordToken],
			m.envValues[fieldGuildID],
			m.envValues[fieldRiotKey],
			provider,
			model,
			m.envValues[fieldAnthropicKey],
			m.envValues[fieldGoogleKey],
		)

		if err := os.WriteFile(".env", []byte(content), 0600); err != nil {
			return stepErrorMsg{err}
		}
		return stepDoneMsg{skipped: false}
	}
}

func (m model) runCurrentStep() tea.Cmd {
	switch m.step {
	case stepDocker:
		return checkDocker()
	case stepAtlas:
		return checkAtlas()
	case stepAir:
		return checkAir()
	case stepEnv:
		return nil
	case stepPostgres:
		return startPostgres()
	case stepSchema:
		return applySchema()
	case stepGitHooks:
		return configureGitHooks()
	case stepComplete:
		return nil
	}
	return nil
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func isDockerRunning() bool {
	err := exec.Command("docker", "info").Run()
	return err == nil
}

func isContainerRunning(name string) bool {
	out, _ := exec.Command("docker", "ps", "-q", "-f", "name="+name).Output()
	return len(strings.TrimSpace(string(out))) > 0
}

func envFileExists() bool {
	_, err := os.Stat(".env")
	return err == nil
}

func checkDocker() tea.Cmd {
	return func() tea.Msg {
		if commandExists("docker") && isDockerRunning() {
			return stepDoneMsg{skipped: true}
		}

		if commandExists("docker") && !isDockerRunning() {
			return dockerReloginMsg{}
		}

		if runtime.GOOS == "darwin" {
			cmd := exec.Command("brew", "install", "--cask", "docker")
			if err := cmd.Run(); err != nil {
				return stepErrorMsg{fmt.Errorf("failed to install Docker: %w", err)}
			}
			return stepDoneMsg{skipped: false}
		}

		cmd := exec.Command("sh", "-c", "curl -fsSL https://get.docker.com | sh")
		if err := cmd.Run(); err != nil {
			return stepErrorMsg{fmt.Errorf("failed to install Docker: %w", err)}
		}

		exec.Command("sudo", "usermod", "-aG", "docker", os.Getenv("USER")).Run()
		return dockerReloginMsg{}
	}
}

func checkAtlas() tea.Cmd {
	return func() tea.Msg {
		if commandExists("atlas") {
			return stepDoneMsg{skipped: true}
		}

		cmd := exec.Command("sh", "-c", "curl -sSf https://atlasgo.sh | sh")
		if err := cmd.Run(); err != nil {
			return stepErrorMsg{fmt.Errorf("failed to install Atlas: %w", err)}
		}
		return stepDoneMsg{skipped: false}
	}
}

func checkAir() tea.Cmd {
	return func() tea.Msg {
		if commandExists("air") {
			return stepDoneMsg{skipped: true}
		}

		cmd := exec.Command("go", "install", "github.com/air-verse/air@latest")
		if err := cmd.Run(); err != nil {
			return stepErrorMsg{fmt.Errorf("failed to install Air: %w", err)}
		}
		return stepDoneMsg{skipped: false}
	}
}

func startPostgres() tea.Cmd {
	return func() tea.Msg {
		if isContainerRunning("leagueofren-db") {
			return stepDoneMsg{skipped: true}
		}

		cmd := exec.Command("docker", "compose", "up", "-d")
		if err := cmd.Run(); err != nil {
			return stepErrorMsg{fmt.Errorf("failed to start PostgreSQL: %w", err)}
		}

		return stepDoneMsg{skipped: false}
	}
}

func applySchema() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("atlas", "schema", "apply", "--env", "local", "--auto-approve")
		if err := cmd.Run(); err != nil {
			return stepErrorMsg{fmt.Errorf("failed to apply schema: %w", err)}
		}
		return stepDoneMsg{skipped: false}
	}
}

func configureGitHooks() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("git", "config", "core.hooksPath", ".githooks")
		if err := cmd.Run(); err != nil {
			return stepErrorMsg{fmt.Errorf("failed to configure git hooks: %w", err)}
		}
		return stepDoneMsg{skipped: false}
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
