package log

import (
	"context"
	"github.com/oumed/titan/internal/app/titan/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// L is the global logger
var L *zap.Logger

// Initialize the global logger
func init() {
	var err error
	L, err = NewLogger(config.C.Log)
	if err != nil {
		panic("Error creating global logger; exception: " + err.Error())
	}
}

// NewLogger creates a new logger
func NewLogger(conf config.Logging) (*zap.Logger, error) {

	c := zap.NewProductionConfig()
	c.Level.UnmarshalText([]byte(conf.Level))
	c.Sampling = nil
	c.Encoding = conf.Encoding
	c.EncoderConfig.EncodeTime = zapcore.EpochNanosTimeEncoder
	c.DisableCaller = !conf.IncludeCaller

	if !conf.IncludeStacktrace {
		// disable stacktrace
		//	 - set the stacktrace key to "" otherwise the stacktrace is still
		//		 shown for log levels above error
		c.DisableStacktrace = true
		c.EncoderConfig.StacktraceKey = ""
	}

	// console is usually used in dev and epoch is meaningless to humans
	// and colour is nice
	if conf.Encoding == "console" {
		c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		c.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	return c.Build()
}

// RequestID returns a zap.Field created from the x-request-id header
func RequestID(ctx context.Context) zap.Field {
	id, ok := ctx.Value("x-request-id").(string)
	if !ok {
		return zap.Skip()
	}
	return zap.String("request_id", id)
}
