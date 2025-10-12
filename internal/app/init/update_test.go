package init

import (
	"errors"
	"sync"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcserv/slacky/internal/config"
	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/tui/styles"
)

var initOnce sync.Once

func setupI18n() {
	initOnce.Do(func() {
		if err := slackyI18n.Init(); err != nil {
			panic(err)
		}
	})
}

// createTestModel creates a minimal test model
func createTestModel() Model {
	setupI18n()

	botInput := textinput.New()
	botInput.Placeholder = "xoxb-..."
	botInput.CharLimit = 200
	botInput.Width = 60

	socketInput := textinput.New()
	socketInput.Placeholder = "xapp-..."
	socketInput.CharLimit = 200
	socketInput.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Label

	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	return Model{
		step:        StepWelcome,
		botToken:    botInput,
		socketToken: socketInput,
		spinner:     s,
		localizer:   localizer,
	}
}

func TestUpdate_ConfigSaved(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepPreferences

	msg := configSavedMsg{}
	result, cmd := update(m, msg)

	resultModel := result.(Model)

	// Should set continueToApp flag
	assert.True(t, resultModel.continueToApp, "continueToApp should be true after config is saved")

	// Should quit to allow transition to main app
	assert.NotNil(t, cmd, "should return quit command")

	// Execute the command to verify it's tea.Quit
	quitMsg := cmd()
	_, isQuitMsg := quitMsg.(tea.QuitMsg)
	assert.True(t, isQuitMsg, "should return tea.Quit message")
}

func TestUpdate_AuthTestSuccess(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepTesting

	msg := authTestMsg{
		teamName: "Test Team",
		userName: "test_user",
		err:      nil,
	}

	result, _ := update(m, msg)
	resultModel := result.(Model)

	assert.Equal(t, StepPreferences, resultModel.step, "should transition to preferences step")
	assert.Equal(t, "Test Team", resultModel.teamName, "should store team name")
	assert.Equal(t, "test_user", resultModel.userName, "should store user name")
	assert.Nil(t, resultModel.err, "should not have error")
}

func TestUpdate_AuthTestFailure(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepTesting

	testErr := errors.New("invalid token")
	msg := authTestMsg{
		err: testErr,
	}

	result, _ := update(m, msg)
	resultModel := result.(Model)

	assert.Equal(t, StepAuthFailed, resultModel.step, "should transition to auth failed step")
	assert.Equal(t, testErr, resultModel.err, "should store error")
}

func TestUpdate_BotTokenTestSuccess(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepBotTokenTesting

	msg := botTokenTestMsg{
		teamName: "Test Team",
		userName: "test_user",
		err:      nil,
	}

	result, cmd := update(m, msg)
	resultModel := result.(Model)

	assert.Equal(t, StepSocketToken, resultModel.step, "should transition to socket token step")
	assert.Equal(t, "Test Team", resultModel.teamName, "should store team name")
	assert.Equal(t, "test_user", resultModel.userName, "should store user name")
	assert.Nil(t, resultModel.err, "should not have error")
	assert.NotNil(t, cmd, "should return command for socket token input blink")
}

func TestUpdate_BotTokenTestFailure(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepBotTokenTesting

	testErr := errors.New("invalid bot token")
	msg := botTokenTestMsg{
		err: testErr,
	}

	result, _ := update(m, msg)
	resultModel := result.(Model)

	assert.Equal(t, StepAuthFailed, resultModel.step, "should transition to auth failed step")
	assert.Equal(t, testErr, resultModel.err, "should store error")
}

func TestUpdate_ErrorMessage(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepBotToken

	testErr := errors.New("config save failed")
	msg := errMsg{err: testErr}

	result, _ := update(m, msg)
	resultModel := result.(Model)

	assert.Equal(t, StepError, resultModel.step, "should transition to error step")
	assert.Equal(t, testErr, resultModel.err, "should store error")
}

func TestUpdate_QuitKey(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepBotToken

	msg := tea.KeyMsg{
		Type: tea.KeyCtrlC,
	}

	result, cmd := update(m, msg)
	resultModel := result.(Model)

	assert.True(t, resultModel.quitting, "quitting flag should be set")
	assert.NotNil(t, cmd, "should return quit command")

	// Execute the command to verify it's tea.Quit
	quitMsg := cmd()
	_, isQuitMsg := quitMsg.(tea.QuitMsg)
	assert.True(t, isQuitMsg, "should return tea.Quit message")
}

func TestHandleEnter_WelcomeStep(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepWelcome

	result, cmd := handleEnter(m)
	resultModel := result.(Model)

	assert.Equal(t, StepBotToken, resultModel.step, "should transition to bot token step")
	assert.NotNil(t, cmd, "should return command for bot token input focus")
}

func TestHandleEnter_BotTokenStep_EmptyToken(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepBotToken
	m.botToken.SetValue("")

	result, _ := handleEnter(m)
	resultModel := result.(Model)

	assert.Equal(t, StepBotToken, resultModel.step, "should stay on bot token step")
	require.NotNil(t, resultModel.err, "should have error")
	assert.Contains(t, resultModel.err.Error(), "required", "error should mention token is required")
}

func TestHandleEnter_BotTokenStep_ValidToken(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepBotToken
	m.botToken.SetValue("xoxb-test-token")

	result, cmd := handleEnter(m)
	resultModel := result.(Model)

	assert.Equal(t, StepBotTokenTesting, resultModel.step, "should transition to testing step")
	assert.Nil(t, resultModel.err, "should clear any previous error")
	assert.NotNil(t, cmd, "should return command to test bot token")
}

func TestHandleEnter_SocketTokenStep_EmptyToken(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepSocketToken
	m.socketToken.SetValue("")

	result, _ := handleEnter(m)
	resultModel := result.(Model)

	assert.Equal(t, StepSocketToken, resultModel.step, "should stay on socket token step")
	require.NotNil(t, resultModel.err, "should have error")
	assert.Contains(t, resultModel.err.Error(), "required", "error should mention token is required")
}

func TestHandleEnter_SocketTokenStep_ValidToken(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepSocketToken
	m.botToken.SetValue("xoxb-test-token")
	m.socketToken.SetValue("xapp-test-token")

	result, cmd := handleEnter(m)
	resultModel := result.(Model)

	assert.Equal(t, StepTesting, resultModel.step, "should transition to testing step")
	assert.Nil(t, resultModel.err, "should clear any previous error")
	assert.NotNil(t, cmd, "should return command to test auth")
}

func TestHandleEnter_PreferencesStep(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepPreferences
	m.botToken.SetValue("xoxb-test-token")
	m.socketToken.SetValue("xapp-test-token")
	m.vimMode = true
	m.showTimestamps = false

	result, cmd := handleEnter(m)
	resultModel := result.(Model)

	// Model should remain the same, but should return save config command
	assert.Equal(t, StepPreferences, resultModel.step, "should stay on preferences step until config is saved")
	assert.NotNil(t, cmd, "should return command to save config")
}

func TestHandleEnter_AuthFailedStep_EditBotToken(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepAuthFailed
	m.authFailCursor = 0
	m.err = errors.New("auth failed")

	result, cmd := handleEnter(m)
	resultModel := result.(Model)

	assert.Equal(t, StepBotToken, resultModel.step, "should transition back to bot token step")
	assert.Nil(t, resultModel.err, "should clear error")
	assert.NotNil(t, cmd, "should return command for bot token input focus")
}

func TestHandleEnter_AuthFailedStep_EditSocketToken(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepAuthFailed
	m.authFailCursor = 1
	m.err = errors.New("auth failed")

	result, cmd := handleEnter(m)
	resultModel := result.(Model)

	assert.Equal(t, StepSocketToken, resultModel.step, "should transition back to socket token step")
	assert.Nil(t, resultModel.err, "should clear error")
	assert.NotNil(t, cmd, "should return command for socket token input focus")
}

func TestHandleEnter_AuthFailedStep_Retry(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepAuthFailed
	m.authFailCursor = 2
	m.err = errors.New("auth failed")
	m.botToken.SetValue("xoxb-test-token")
	m.socketToken.SetValue("xapp-test-token")

	result, cmd := handleEnter(m)
	resultModel := result.(Model)

	assert.Equal(t, StepTesting, resultModel.step, "should transition to testing step")
	assert.Nil(t, resultModel.err, "should clear error")
	assert.NotNil(t, cmd, "should return command to test auth")
}

func TestHandleEnter_ErrorStep(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepError
	m.err = errors.New("some error")

	result, cmd := handleEnter(m)
	resultModel := result.(Model)

	assert.True(t, resultModel.quitting, "quitting flag should be set")
	assert.NotNil(t, cmd, "should return quit command")

	// Execute the command to verify it's tea.Quit
	quitMsg := cmd()
	_, isQuitMsg := quitMsg.(tea.QuitMsg)
	assert.True(t, isQuitMsg, "should return tea.Quit message")
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.width = 0

	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}

	result, _ := update(m, msg)
	resultModel := result.(Model)

	assert.Equal(t, 120, resultModel.width, "should update width")
	assert.NotEmpty(t, resultModel.logoRendered, "should render logo")
}

func TestUpdate_ToggleKey_Preferences(t *testing.T) {
	t.Parallel()

	m := createTestModel()
	m.step = StepPreferences
	m.showTimestamps = false
	m.prefCursor = 0

	msg := tea.KeyMsg{
		Type: tea.KeySpace,
	}

	result, _ := update(m, msg)
	resultModel := result.(Model)

	assert.True(t, resultModel.showTimestamps, "should toggle timestamps preference")
}

func TestUpdate_NavigationKeys_AuthFailed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialCursor  int
		keyType        tea.KeyType
		expectedCursor int
	}{
		{
			name:           "Down from 0",
			initialCursor:  0,
			keyType:        tea.KeyDown,
			expectedCursor: 1,
		},
		{
			name:           "Down from 1",
			initialCursor:  1,
			keyType:        tea.KeyDown,
			expectedCursor: 2,
		},
		{
			name:           "Down from 2 (max)",
			initialCursor:  2,
			keyType:        tea.KeyDown,
			expectedCursor: 2,
		},
		{
			name:           "Up from 2",
			initialCursor:  2,
			keyType:        tea.KeyUp,
			expectedCursor: 1,
		},
		{
			name:           "Up from 1",
			initialCursor:  1,
			keyType:        tea.KeyUp,
			expectedCursor: 0,
		},
		{
			name:           "Up from 0 (min)",
			initialCursor:  0,
			keyType:        tea.KeyUp,
			expectedCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := createTestModel()
			m.step = StepAuthFailed
			m.authFailCursor = tt.initialCursor

			msg := tea.KeyMsg{
				Type: tt.keyType,
			}

			result, _ := update(m, msg)
			resultModel := result.(Model)

			assert.Equal(t, tt.expectedCursor, resultModel.authFailCursor, "cursor should be at expected position")
		})
	}
}

// TestInitFlowComplete tests the complete init flow to main app transition
func TestInitFlowComplete(t *testing.T) {
	t.Parallel()

	// Start with initial model
	m := createTestModel()

	// Step 1: Welcome -> Bot Token
	result, _ := handleEnter(m)
	m = result.(Model)
	assert.Equal(t, StepBotToken, m.step, "should be on bot token step")

	// Step 2: Enter bot token and test
	m.botToken.SetValue("xoxb-test-token")
	result, _ = handleEnter(m)
	m = result.(Model)
	assert.Equal(t, StepBotTokenTesting, m.step, "should be testing bot token")

	// Step 3: Bot token test succeeds
	result, _ = update(m, botTokenTestMsg{
		teamName: "Test Team",
		userName: "test_user",
	})
	m = result.(Model)
	assert.Equal(t, StepSocketToken, m.step, "should be on socket token step")

	// Step 4: Enter socket token and test
	m.socketToken.SetValue("xapp-test-token")
	result, _ = handleEnter(m)
	m = result.(Model)
	assert.Equal(t, StepTesting, m.step, "should be testing auth")

	// Step 5: Auth test succeeds
	result, _ = update(m, authTestMsg{
		teamName: "Test Team",
		userName: "test_user",
	})
	m = result.(Model)
	assert.Equal(t, StepPreferences, m.step, "should be on preferences step")

	// Step 6: Complete preferences and save config
	result, _ = handleEnter(m)
	m = result.(Model)
	// At this point, saveConfig command is returned but we can't easily test it
	// So we'll simulate the configSavedMsg being received

	// Step 7: Config saved successfully
	result, cmd := update(m, configSavedMsg{})
	finalModel := result.(Model)

	// CRITICAL: After config is saved, continueToApp should be true
	// and the program should quit to allow transition to main app
	assert.True(t, finalModel.continueToApp, "continueToApp should be true after successful init")
	assert.NotNil(t, cmd, "should return quit command to transition to main app")

	// Verify quit command
	quitMsg := cmd()
	_, isQuitMsg := quitMsg.(tea.QuitMsg)
	assert.True(t, isQuitMsg, "should return tea.Quit to allow main app to start")
}

func TestInitModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		existingConfig *config.Config
		expectedVim    bool
		expectedTS     bool
	}{
		{
			name:           "No existing config",
			existingConfig: nil,
			expectedVim:    true,
			expectedTS:     true,
		},
		{
			name: "Existing config with custom preferences",
			existingConfig: &config.Config{
				UI: config.UI{
					VimMode:        false,
					ShowTimestamps: false,
				},
			},
			expectedVim: false,
			expectedTS:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := initialModel(tt.existingConfig)

			assert.Equal(t, StepWelcome, m.step, "should start at welcome step")
			assert.Equal(t, tt.expectedVim, m.vimMode, "vim mode should match expected")
			assert.Equal(t, tt.expectedTS, m.showTimestamps, "show timestamps should match expected")
			assert.NotNil(t, m.botToken, "bot token input should be initialized")
			assert.NotNil(t, m.socketToken, "socket token input should be initialized")
			assert.NotNil(t, m.spinner, "spinner should be initialized")
			assert.NotNil(t, m.localizer, "localizer should be initialized")
		})
	}
}

func TestUpdate_SpinnerTick(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		step         Step
		shouldUpdate bool
	}{
		{
			name:         "Spinner updates during bot token testing",
			step:         StepBotTokenTesting,
			shouldUpdate: true,
		},
		{
			name:         "Spinner updates during testing",
			step:         StepTesting,
			shouldUpdate: true,
		},
		{
			name:         "Spinner doesn't update on other steps",
			step:         StepBotToken,
			shouldUpdate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := createTestModel()
			m.step = tt.step

			// Create a spinner tick message
			tickCmd := m.spinner.Tick
			tickMsg := tickCmd()

			result, cmd := update(m, tickMsg)

			if tt.shouldUpdate {
				assert.NotNil(t, cmd, "should return spinner tick command")
			}
			assert.NotNil(t, result, "should return updated model")
		})
	}
}

func TestKeys_AreProperlyDefined(t *testing.T) {
	t.Parallel()

	// Verify that all keys used in update.go are properly defined
	assert.True(t, key.Matches(tea.KeyMsg{Type: tea.KeyCtrlC}, keys.Quit), "Ctrl+C should match Quit key")
	assert.True(t, key.Matches(tea.KeyMsg{Type: tea.KeyEnter}, keys.Enter), "Enter should match Enter key")
	assert.True(t, key.Matches(tea.KeyMsg{Type: tea.KeyUp}, keys.Up), "Up arrow should match Up key")
	assert.True(t, key.Matches(tea.KeyMsg{Type: tea.KeyDown}, keys.Down), "Down arrow should match Down key")
	assert.True(t, key.Matches(tea.KeyMsg{Type: tea.KeySpace}, keys.Toggle), "Space should match Toggle key")
}
