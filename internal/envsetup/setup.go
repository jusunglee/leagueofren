// envsetup provides a lightweight .env configuration wizard.
// It runs automatically on first bot startup when no .env file exists,
// collecting Discord, Riot, and LLM credentials.
package envsetup

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type step int

const (
	stepWelcome step = iota
	stepDiscord
	stepRiot
	stepLLMProvider
	stepLLMKey
	stepConfirm
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Underline(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

type model struct {
	step         step
	discordToken string
	riotAPIKey   string
	llmProvider  string
	llmAPIKey    string
	input        string
	err          error
	width        int
	height       int
}

func New() model {
	return model{
		step: stepWelcome,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyBackspace:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			return m, nil

		case tea.KeyRunes:
			m.input += string(msg.Runes)
			return m, nil

		case tea.KeySpace:
			m.input += " "
			return m, nil
		}
	}

	return m, nil
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	m.err = nil

	switch m.step {
	case stepWelcome:
		m.step = stepDiscord
		m.input = ""

	case stepDiscord:
		token := strings.TrimSpace(m.input)
		if token == "" {
			m.err = fmt.Errorf("Discord token is required")
			return m, nil
		}
		m.discordToken = token
		m.step = stepRiot
		m.input = ""

	case stepRiot:
		key := strings.TrimSpace(m.input)
		if key == "" {
			m.err = fmt.Errorf("Riot API key is required")
			return m, nil
		}
		m.riotAPIKey = key
		m.step = stepLLMProvider
		m.input = ""

	case stepLLMProvider:
		choice := strings.TrimSpace(strings.ToLower(m.input))
		if choice != "1" && choice != "2" && choice != "anthropic" && choice != "google" {
			m.err = fmt.Errorf("Please enter 1 for Anthropic or 2 for Google")
			return m, nil
		}
		if choice == "1" || choice == "anthropic" {
			m.llmProvider = "anthropic"
		} else {
			m.llmProvider = "google"
		}
		m.step = stepLLMKey
		m.input = ""

	case stepLLMKey:
		key := strings.TrimSpace(m.input)
		if key == "" {
			m.err = fmt.Errorf("API key is required")
			return m, nil
		}
		m.llmAPIKey = key
		m.step = stepConfirm
		m.input = ""

	case stepConfirm:
		choice := strings.TrimSpace(strings.ToLower(m.input))
		if choice == "y" || choice == "yes" || choice == "" {
			if err := m.writeEnvFile(); err != nil {
				m.err = err
				return m, nil
			}
			return m, tea.Quit
		} else if choice == "n" || choice == "no" {
			m.step = stepWelcome
			m.input = ""
			m.discordToken = ""
			m.riotAPIKey = ""
			m.llmProvider = ""
			m.llmAPIKey = ""
		}
	}

	return m, nil
}

func (m model) writeEnvFile() error {
	var llmModel string
	var llmKeyName string
	if m.llmProvider == "anthropic" {
		llmModel = "claude-sonnet-4-20250514"
		llmKeyName = "ANTHROPIC_API_KEY"
	} else {
		llmModel = "gemini-2.0-flash"
		llmKeyName = "GOOGLE_API_KEY"
	}

	content := fmt.Sprintf(`DATABASE_URL=./leagueofren.db
DISCORD_TOKEN=%s
RIOT_API_KEY=%s
LLM_PROVIDER=%s
LLM_MODEL=%s
%s=%s
`, m.discordToken, m.riotAPIKey, m.llmProvider, llmModel, llmKeyName, m.llmAPIKey)

	return os.WriteFile(".env", []byte(content), 0600)
}

func (m model) View() string {
	var s strings.Builder

	switch m.step {
	case stepWelcome:
		s.WriteString(titleStyle.Render("League of Ren - Env Setup"))
		s.WriteString("\n\n")
		s.WriteString("This wizard will help you configure the bot.\n")
		s.WriteString("You'll need:\n\n")
		s.WriteString("  - A Discord bot token\n")
		s.WriteString("  - A Riot Games API key\n")
		s.WriteString("  - An LLM API key (Anthropic or Google)\n")
		s.WriteString("\n")
		s.WriteString(dimStyle.Render("Press Enter to continue, Ctrl+C to exit"))

	case stepDiscord:
		s.WriteString(titleStyle.Render("Step 1: Discord Bot Token"))
		s.WriteString("\n\n")
		s.WriteString("To get your Discord bot token:\n\n")
		s.WriteString("  1. Go to " + linkStyle.Render("https://discord.com/developers/applications") + "\n")
		s.WriteString("  2. Create a new application (or select existing)\n")
		s.WriteString("  3. Go to the Bot section\n")
		s.WriteString("  4. Click 'Reset Token' to get your bot token\n")
		s.WriteString("  5. Enable 'Message Content Intent' under Privileged Gateway Intents\n")
		s.WriteString("\n")
		s.WriteString(labelStyle.Render("Paste your Discord token here:"))
		s.WriteString("\n")
		s.WriteString("> " + inputStyle.Render(maskToken(m.input)))
		if m.err != nil {
			s.WriteString("\n" + errorStyle.Render(m.err.Error()))
		}

	case stepRiot:
		s.WriteString(titleStyle.Render("Step 2: Riot Games API Key"))
		s.WriteString("\n\n")
		s.WriteString("To get your Riot API key:\n\n")
		s.WriteString("  1. Go to " + linkStyle.Render("https://developer.riotgames.com") + "\n")
		s.WriteString("  2. Sign in with your Riot account\n")
		s.WriteString("  3. Your development API key is shown on the dashboard\n")
		s.WriteString("  4. For production, register your app for a production key\n")
		s.WriteString("\n")
		s.WriteString(labelStyle.Render("Paste your Riot API key here:"))
		s.WriteString("\n")
		s.WriteString("> " + inputStyle.Render(maskToken(m.input)))
		if m.err != nil {
			s.WriteString("\n" + errorStyle.Render(m.err.Error()))
		}

	case stepLLMProvider:
		s.WriteString(titleStyle.Render("Step 3: Choose LLM Provider"))
		s.WriteString("\n\n")
		s.WriteString("Which LLM provider would you like to use?\n\n")
		s.WriteString("  1. Anthropic (Claude)\n")
		s.WriteString("  2. Google (Gemini)\n")
		s.WriteString("\n")
		s.WriteString(labelStyle.Render("Enter 1 or 2:"))
		s.WriteString("\n")
		s.WriteString("> " + inputStyle.Render(m.input))
		if m.err != nil {
			s.WriteString("\n" + errorStyle.Render(m.err.Error()))
		}

	case stepLLMKey:
		s.WriteString(titleStyle.Render("Step 4: LLM API Key"))
		s.WriteString("\n\n")
		if m.llmProvider == "anthropic" {
			s.WriteString("To get your Anthropic API key:\n\n")
			s.WriteString("  1. Go to " + linkStyle.Render("https://console.anthropic.com") + "\n")
			s.WriteString("  2. Sign up or log in\n")
			s.WriteString("  3. Go to API Keys and create a new key\n")
		} else {
			s.WriteString("To get your Google AI API key:\n\n")
			s.WriteString("  1. Go to " + linkStyle.Render("https://aistudio.google.com/apikey") + "\n")
			s.WriteString("  2. Sign in with your Google account\n")
			s.WriteString("  3. Create an API key\n")
		}
		s.WriteString("\n")
		s.WriteString(labelStyle.Render("Paste your API key here:"))
		s.WriteString("\n")
		s.WriteString("> " + inputStyle.Render(maskToken(m.input)))
		if m.err != nil {
			s.WriteString("\n" + errorStyle.Render(m.err.Error()))
		}

	case stepConfirm:
		s.WriteString(titleStyle.Render("Configuration Complete"))
		s.WriteString("\n\n")
		s.WriteString("Your configuration:\n\n")
		s.WriteString("  Database:     " + successStyle.Render("./leagueofren.db") + "\n")
		s.WriteString("  Discord:      " + successStyle.Render(maskToken(m.discordToken)) + "\n")
		s.WriteString("  Riot API:     " + successStyle.Render(maskToken(m.riotAPIKey)) + "\n")
		s.WriteString("  LLM Provider: " + successStyle.Render(m.llmProvider) + "\n")
		s.WriteString("  LLM API Key:  " + successStyle.Render(maskToken(m.llmAPIKey)) + "\n")
		s.WriteString("\n")
		s.WriteString(labelStyle.Render("Save this configuration? [Y/n]:"))
		s.WriteString("\n")
		s.WriteString("> " + inputStyle.Render(m.input))
		if m.err != nil {
			s.WriteString("\n" + errorStyle.Render(m.err.Error()))
		}
	}

	s.WriteString("\n")
	return s.String()
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}

// Run starts the setup wizard and returns true if setup was completed successfully
func Run() (bool, error) {
	p := tea.NewProgram(New())
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	m := finalModel.(model)
	return m.step == stepConfirm && m.err == nil, nil
}

// NeedsSetup checks if .env file exists
func NeedsSetup() bool {
	_, err := os.Stat(".env")
	return os.IsNotExist(err)
}
