package log

import "log"

var Logger LoggerInterface

func init() {
	Logger = &DefaultLogger{
		logger: log.Default(),
	}
}

func InitLogger(ILogger LoggerInterface) {
	Logger = ILogger
}

type LoggerInterface interface {
	Debugf(format string, params ...interface{})
	Infof(format string, params ...interface{})
	Warnf(format string, params ...interface{})
	Errorf(format string, params ...interface{})

	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
}

func Debugf(format string, params ...interface{}) {
	Logger.Debugf(format, params...)
}

func Infof(format string, params ...interface{}) {
	Logger.Debugf(format, params...)
}

func Warnf(format string, params ...interface{}) {
	Logger.Warnf(format, params...)
}

func Errorf(format string, params ...interface{}) {
	Logger.Errorf(format, params...)
}

func Debug(v ...interface{}) {
	Logger.Debug(v...)
}

func Info(v ...interface{}) {
	Logger.Info(v...)
}

func Warn(v ...interface{}) {
	Logger.Warn(v...)
}

func Error(v ...interface{}) {
	Logger.Error(v...)
}

type DefaultLogger struct {
	logger *log.Logger
}

// Errorf implements LoggerInterface.
func (d *DefaultLogger) Errorf(format string, params ...interface{}) {
	d.logger.Printf(format, params...)
}

// Infof implements LoggerInterface.
func (d *DefaultLogger) Infof(format string, params ...interface{}) {
	d.logger.Printf(format, params...)
}

// Warnf implements LoggerInterface.
func (d *DefaultLogger) Warnf(format string, params ...interface{}) {
	d.logger.Printf(format, params...)
}

func (d *DefaultLogger) Debugf(format string, params ...interface{}) {
	d.logger.Printf(format, params...)
}

// Debug implements LoggerInterface.
func (d *DefaultLogger) Debug(v ...interface{}) {
	d.logger.Print(v...)
}

// Info implements LoggerInterface.
func (d *DefaultLogger) Info(v ...interface{}) {
	d.logger.Print(v...)
}

// Warn implements LoggerInterface.
func (d *DefaultLogger) Warn(v ...interface{}) {
	d.logger.Print(v...)
}

// Error implements LoggerInterface.
func (d *DefaultLogger) Error(v ...interface{}) {
	d.logger.Print(v...)
}
