package whatsapp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	waLog "go.mau.fi/whatsmeow/util/log"
)

type LoggingOptions struct {
	DebugLogDir        string
	DebugLogPerAccount bool
	DebugLogLevel      string
	ConsoleLogLevel    string
}

var defaultLoggingOptions = LoggingOptions{
	DebugLogDir:        "debug_log",
	DebugLogPerAccount: true,
	DebugLogLevel:      "DEBUG",
	ConsoleLogLevel:    "INFO",
}

var (
	loggingOptionsMu sync.RWMutex
	loggingOptions   = defaultLoggingOptions
)

var levelToInt = map[string]int{
	"":      -1,
	"DEBUG": 0,
	"INFO":  1,
	"WARN":  2,
	"ERROR": 3,
}

func ConfigureLogging(opts LoggingOptions) {
	if strings.TrimSpace(opts.DebugLogDir) == "" {
		opts.DebugLogDir = defaultLoggingOptions.DebugLogDir
	}
	if strings.TrimSpace(opts.DebugLogLevel) == "" {
		opts.DebugLogLevel = defaultLoggingOptions.DebugLogLevel
	}
	if strings.TrimSpace(opts.ConsoleLogLevel) == "" {
		opts.ConsoleLogLevel = defaultLoggingOptions.ConsoleLogLevel
	}

	opts.DebugLogLevel = strings.ToUpper(strings.TrimSpace(opts.DebugLogLevel))
	opts.ConsoleLogLevel = strings.ToUpper(strings.TrimSpace(opts.ConsoleLogLevel))

	loggingOptionsMu.Lock()
	loggingOptions = opts
	loggingOptionsMu.Unlock()
}

func getLoggingOptions() LoggingOptions {
	loggingOptionsMu.RLock()
	defer loggingOptionsMu.RUnlock()
	return loggingOptions
}

type dualLogger struct {
	mod        string
	accountID  string
	consoleMin int
	fileMin    int
}

func NewModuleLogger(module string) waLog.Logger {
	module = strings.TrimSpace(module)
	if module == "" {
		module = "WA"
	}
	opts := getLoggingOptions()
	return &dualLogger{
		mod:        module,
		accountID:  "",
		consoleMin: levelToInt[opts.ConsoleLogLevel],
		fileMin:    levelToInt[opts.DebugLogLevel],
	}
}

func NewWhatsAppClientLogger(accountID string) waLog.Logger {
	accountID = NormalizeAccountID(accountID)
	opts := getLoggingOptions()
	return &dualLogger{
		mod:        "WA-" + accountID,
		accountID:  accountID,
		consoleMin: levelToInt["WARN"],
		fileMin:    levelToInt[opts.DebugLogLevel],
	}
}

func NewWhatsAppDBLogger(accountID string) waLog.Logger {
	accountID = NormalizeAccountID(accountID)
	opts := getLoggingOptions()
	return &dualLogger{
		mod:        "DB-" + accountID,
		accountID:  accountID,
		consoleMin: levelToInt["WARN"],
		fileMin:    levelToInt[opts.DebugLogLevel],
	}
}

func (l *dualLogger) Errorf(msg string, args ...interface{}) { l.outputf("ERROR", msg, args...) }
func (l *dualLogger) Warnf(msg string, args ...interface{})  { l.outputf("WARN", msg, args...) }
func (l *dualLogger) Infof(msg string, args ...interface{})  { l.outputf("INFO", msg, args...) }
func (l *dualLogger) Debugf(msg string, args ...interface{}) { l.outputf("DEBUG", msg, args...) }

func (l *dualLogger) Sub(module string) waLog.Logger {
	mod := strings.TrimSpace(module)
	if mod == "" {
		mod = "sub"
	}
	return &dualLogger{
		mod:        fmt.Sprintf("%s/%s", l.mod, mod),
		accountID:  l.accountID,
		consoleMin: l.consoleMin,
		fileMin:    l.fileMin,
	}
}

func (l *dualLogger) outputf(level, msg string, args ...interface{}) {
	level = strings.ToUpper(strings.TrimSpace(level))
	if level == "" {
		level = "INFO"
	}
	levelInt := levelToInt[level]
	line := fmt.Sprintf("%s [%s %s] %s", time.Now().Format("15:04:05.000"), l.mod, level, fmt.Sprintf(msg, args...))

	if levelInt >= l.consoleMin {
		fmt.Println(line)
	}
	if levelInt >= l.fileMin {
		writeDebugLine(l.accountID, line)
	}
}

func writeDebugLine(accountID, line string) {
	opts := getLoggingOptions()
	baseDir := strings.TrimSpace(opts.DebugLogDir)
	if baseDir == "" {
		baseDir = defaultLoggingOptions.DebugLogDir
	}

	date := time.Now().Format("2006-01-02")
	dir := filepath.Join(baseDir, date)
	fileName := "whatsapp-debug.log"

	if opts.DebugLogPerAccount {
		accountID = NormalizeAccountID(accountID)
		if accountID != "" {
			dir = filepath.Join(dir, "account-"+sanitizeLogPathFragment(accountID))
			fileName = "whatsapp-debug.log"
		}
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}

	path := filepath.Join(dir, fileName)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.WriteString(line + "\n")
}

func sanitizeLogPathFragment(value string) string {
	if value == "" {
		return "default"
	}
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('_')
	}
	result := strings.Trim(b.String(), "_")
	if result == "" {
		return "default"
	}
	return result
}
