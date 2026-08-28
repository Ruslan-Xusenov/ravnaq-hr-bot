package worker

import (
	"log/slog"

	"github.com/hibiken/asynq"
)

type WorkerServer struct {
	srv *asynq.Server
	mux *asynq.ServeMux
}

func NewWorkerServer(redisURL string, processor *BroadcastProcessor) *WorkerServer {
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		slog.Error("Failed to parse redis url for asynq", "error", err)
		// Fallback to localhost if parse fails or is empty
		opt = asynq.RedisClientOpt{Addr: "localhost:6379"}
	}

	srv := asynq.NewServer(
		opt,
		asynq.Config{
			Concurrency: 20, // Increased for maximum message delivery speed
			Queues: map[string]int{
				"default": 10,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeBroadcastMessage, processor.HandleBroadcastMessage)

	return &WorkerServer{
		srv: srv,
		mux: mux,
	}
}

func (s *WorkerServer) Start() error {
	slog.Info("Starting Asynq worker server...")
	return s.srv.Start(s.mux)
}

func (s *WorkerServer) Stop() {
	slog.Info("Shutting down Asynq worker server...")
	s.srv.Shutdown()
}
