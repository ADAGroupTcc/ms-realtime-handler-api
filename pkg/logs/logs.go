package logs

import (
	"encoding/json"

	logger "github.com/PicPay/lib-go-logger/v2"
)

func New(formatter logger.Formatter) *logger.Logger {
	return logger.New(logger.WithMinimalLevel(logger.DEBUG), logger.WithFormatter(formatter))
}

type JSONFormatter struct {
	requestId string
}

func (formatter *JSONFormatter) Format(log *logger.Log) []byte {
	log.XRequestID = formatter.requestId
	data, err := json.Marshal(log)
	if err != nil {
		data, _ = json.Marshal(logger.Log{
			Level:     logger.ERROR,
			Type:      log.Type,
			Timestamp: log.Timestamp,
			Message:   err.Error(),
			Event:     log.Event,
		})
	}
	return data
}

func (formatter *JSONFormatter) SetRequestID(requestId string) {
	formatter.requestId = requestId
}
