package resources

import (
	"os"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

var (
	Log           log.Logger
	KafkaProducer sarama.SyncProducer
)

func init() {
	logLevel := os.Getenv("loglevel")
	logger := log.NewJSONLogger(os.Stderr)
	logger = addLogFilter(logger, logLevel)
	logger = log.With(logger, "time", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	Log = logger

	initializeProducer()
}

func addLogFilter(logger log.Logger, logLevel string) log.Logger {
	switch logLevel {
	case "debug":
		return level.NewFilter(logger, level.AllowDebug())
	case "info":
		return level.NewFilter(logger, level.AllowInfo())
	case "warn":
		return level.NewFilter(logger, level.AllowWarn())
	case "error":
		return level.NewFilter(logger, level.AllowError())
	case "none":
		return level.NewFilter(logger, level.AllowNone())
	default:
		return level.NewFilter(logger, level.AllowAll())
	}
}

func initializeProducer() {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Retry.Max = 3
	kafkaConfig.Producer.Return.Successes = true

	bootstrapServers := os.Getenv("bootstrapServers")
	level.Info(Log).Log("bootstrap servers", bootstrapServers)
	client, err := sarama.NewClient(strings.Split(bootstrapServers, ","), kafkaConfig)
	if err != nil {
		return
	}
	level.Info(Log).Log("msg", "connection successfull")

	KafkaProducer, err = sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return
	}
}
