package logger

import (
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	KeyLevelError = "ERROR"
	KeyLevelInfo  = "INFO"
	KeyLevelDebug = "DEBUG"

	EncodingJSON    = "json"
	EncodingConsole = "console"

	TimeKey    = "time"
	MessageKey = "message"
	LevelKey   = "level"
	CallerKey  = "caller"

	ConsoleTimeFormat = "2006-01-02T15:04:05.000"
)

func init() {
	// default unregistered
	MustSetupGlobal(
		WithConfiguration(CoreOptions{
			OutputPath: "stderr",
			Level:      KeyLevelDebug,
			Encoding:   EncodingConsole,
		}),
	)
}

type Logger struct {
	*zap.Logger
	SugaredLogger *zap.SugaredLogger
	ErrorLogger   *zap.Logger
}

type RotateOptions struct {
	MaxSize    int // megabytes
	MaxBackups int
	MaxAge     int // days
}

type CoreOptions struct {
	OutputPath string
	Level      string
	Encoding   string
	TimeFormat string
	Rotate     *RotateOptions
}

type Configuration func(l *Logger) error

type Options struct {
	Level string
	Debug bool
}

// _dLogger defaultLogger
var _dLogger Logger

var (
	ErrWrongLogLevelConfiguration = func(opts ...interface{}) error {
		return fmt.Errorf("wrong logger level: %s, instead [%s]", opts...)
	}
	ErrWrongLogEncodingConfiguration = func(opts ...interface{}) error {
		return fmt.Errorf("wrong logger encoding: %s, instead [%s]", opts...)
	}
)

func New(cfgs ...Configuration) (Logger, error) {
	l := Logger{}

	for _, cfg := range cfgs {
		err := cfg(&l)
		if err != nil {
			return l, err
		}
	}

	return l, nil
}

func MustSetupGlobal(cfgs ...Configuration) Logger {
	l, err := New(cfgs...)
	if err != nil {
		log.Fatalf("Unable create global logger, error: %s", err.Error())
	}

	_dLogger = l

	return l
}

func WithConfiguration(o CoreOptions) Configuration {
	return func(l *Logger) error {
		var (
			err           error
			encoder       zapcore.Encoder
			encoderConfig zapcore.EncoderConfig
			wr            zapcore.WriteSyncer
		)

		dLevel, err := zapcore.ParseLevel(o.Level)
		if err != nil {
			log.Println(
				ErrWrongLogLevelConfiguration(
					o.Level,
					[]string{KeyLevelError, KeyLevelInfo, KeyLevelDebug},
				),
			)

			return err
		}

		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

		if o.TimeFormat != "" {
			encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(o.TimeFormat)
		}

		encoderConfig.LevelKey = LevelKey
		encoderConfig.TimeKey = TimeKey
		encoderConfig.MessageKey = MessageKey
		encoderConfig.CallerKey = CallerKey

		switch o.Encoding {
		case EncodingConsole:
			encoderConfig.EncodeLevel = capitalColorLevelEncoder
			encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
			encoderConfig.EncodeTime = TimeEncoderOfLayout(ConsoleTimeFormat)
			encoderConfig.EncodeCaller = ShortCallerEncoder
			encoderConfig.MessageKey = "\t-\t"

			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		case EncodingJSON:
			encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
			encoderConfig.EncodeDuration = zapcore.MillisDurationEncoder

			encoder = zapcore.NewJSONEncoder(encoderConfig)
		default:
			return ErrWrongLogEncodingConfiguration(o.Encoding, []string{EncodingJSON, EncodingConsole})
		}

		switch o.OutputPath {
		case "stdout":
			wr = zapcore.AddSync(os.Stdout)
		case "stderr":
			wr = zapcore.AddSync(os.Stderr)
		default:
			if o.Rotate != nil {
				wr = zapcore.AddSync(&lumberjack.Logger{
					Filename:   o.OutputPath,
					MaxSize:    o.Rotate.MaxSize,
					MaxBackups: o.Rotate.MaxBackups,
					MaxAge:     o.Rotate.MaxAge,
				})

				break
			}

			wr = zapcore.AddSync(&lumberjack.Logger{
				Filename: o.OutputPath,
			})
		}

		core := zapcore.NewCore(encoder, wr, dLevel)

		if l.Logger != nil {
			core = zapcore.NewTee(
				l.Logger.Core(),
				zapcore.NewCore(encoder, wr, dLevel),
			)
		}

		l.Logger = zap.New(core, zap.WithCaller(true))
		l.SugaredLogger = l.Logger.Sugar()
		l.ErrorLogger = zap.New(l.Logger.Core(), zap.AddCallerSkip(1), zap.WithCaller(true))

		zap.L()

		return nil
	}
}

func (l *Logger) WithError(msg string, err error, fields ...zap.Field) error {
	fields = append(fields, zap.Error(err), zap.Stack("stacktrace"))
	l.ErrorLogger.Error(msg, fields...)

	return err
}

func L() *Logger {
	return &_dLogger
}

func S() *zap.SugaredLogger {
	return _dLogger.SugaredLogger
}

func CurrentDefault() Logger {
	return _dLogger
}
