package main

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"net"
)

const configFilename string = "socks5.yaml"

func main() {

	var cfg config
	var ok bool
	if cfg, ok = tryParseConfig(); !ok {
		return
	}

	logger := newLogger("socks5")
	defer logger.Sync()

	listener, err := Listen(cfg.Network, fmt.Sprintf("%s:%d",cfg.Address,cfg.Port))
	if err != nil {
		logger.Error(err.Error())
		return
	}

	for {
		conn, err := NewConnection(listener)
		if err != nil {
			logger.Error(err.Error())
			break
		}
		logger.Info(fmt.Sprintf("Opened connection from: %v", conn.RemoteAddr()))
		go Start(conn, cfg, logger)
	}

}

func NewConnection(listen net.Listener) (net.Conn, error) {
	conn, err := listen.Accept()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func Listen(network string, addr string) (net.Listener, error) {
	if addr == "" {
		return nil, errors.New("Address is empty")
	}

	listener, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func newLogger(name string) *zap.Logger {

	mainLogger := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/%s.log", "./", name),
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})

	encoderConfig := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey: "message",

		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,

		TimeKey:    "time",
		EncodeTime: zapcore.ISO8601TimeEncoder,

		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,

		EncodeDuration: zapcore.StringDurationEncoder,
	})

	return zap.New(zapcore.NewCore(encoderConfig, mainLogger, zapcore.DebugLevel))
}
