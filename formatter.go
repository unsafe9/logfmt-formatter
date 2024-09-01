package logfmt_formatter

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime"
	"sort"
	"strconv"
)

const (
	RFC3339Milli = "2006-01-02T15:04:05.000Z07:00"
	RFC3339Micro = "2006-01-02T15:04:05.000000Z07:00"
)

const (
	ansiRed    = 31
	ansiYellow = 33
	ansiBlue   = 36
	ansiGray   = 37
)

type Formatter struct {
	DisableColors bool

	DisableSorting bool
	SortingFunc    func([]string)

	ForceQuote       bool
	DisableQuote     bool
	QuoteEmptyFields bool

	DisableTimestamp bool
	TimestampFormat  string

	CallerPrettyfier func(*runtime.Frame) (function string, file string)
}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	var funcVal, fileVal string

	fixedKeys := make([]string, 0, 4+len(data))

	if !f.DisableTimestamp {
		fixedKeys = append(fixedKeys, logrus.FieldKeyTime)
		timestampFormat := f.TimestampFormat
		if timestampFormat == "" {
			timestampFormat = RFC3339Milli
		}
		data[logrus.FieldKeyTime] = entry.Time.Format(timestampFormat)
	}

	fixedKeys = append(fixedKeys, logrus.FieldKeyLevel)
	data[logrus.FieldKeyLevel] = entry.Level.String()

	if entry.Message != "" {
		fixedKeys = append(fixedKeys, logrus.FieldKeyMsg)
		data[logrus.FieldKeyMsg] = entry.Message
	}

	if entry.HasCaller() {
		if f.CallerPrettyfier != nil {
			funcVal, fileVal = f.CallerPrettyfier(entry.Caller)
		} else {
			funcVal = fmt.Sprintf("%s()", entry.Caller.Function)
			fileVal = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		}

		if funcVal != "" {
			fixedKeys = append(fixedKeys, logrus.FieldKeyFunc)
			data[logrus.FieldKeyFunc] = funcVal
		}
		if fileVal != "" {
			fixedKeys = append(fixedKeys, logrus.FieldKeyFile)
			data[logrus.FieldKeyFile] = fileVal
		}
	}

	if !f.DisableSorting {
		if f.SortingFunc == nil {
			sort.Strings(keys)
			fixedKeys = append(fixedKeys, keys...)
		} else {
			fixedKeys = append(fixedKeys, keys...)
			f.SortingFunc(fixedKeys)
		}
	} else {
		fixedKeys = append(fixedKeys, keys...)
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	for _, key := range fixedKeys {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		f.appendKey(b, key, entry.Level)
		b.WriteByte('=')
		f.appendValue(b, data[key])
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *Formatter) isColored() bool {
	return !f.DisableColors
}

func (f *Formatter) needsQuoting(text string) bool {
	if f.ForceQuote {
		return true
	}
	if f.QuoteEmptyFields && len(text) == 0 {
		return true
	}
	if f.DisableQuote {
		return false
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}

func (f *Formatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	f.appendValue(b, value)
}

func (f *Formatter) ansiColorByLevel(level logrus.Level) int {
	switch level {
	case logrus.DebugLevel, logrus.TraceLevel:
		return ansiGray
	case logrus.WarnLevel:
		return ansiYellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return ansiRed
	case logrus.InfoLevel:
		return ansiBlue
	default:
		return ansiBlue
	}
}

func (f *Formatter) appendKey(b *bytes.Buffer, key string, level logrus.Level) {
	if f.isColored() {
		b.WriteString("\x1b[")
		b.WriteString(strconv.Itoa(f.ansiColorByLevel(level)))
		b.WriteString("m")
		b.WriteString(key)
		b.WriteString("\x1b[0m")

	} else {
		b.WriteString(key)
	}
}

func (f *Formatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	if !f.needsQuoting(stringVal) {
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}
