package main

import (
	"context"
	"fmt"
	stdhttp "net/http"
	"os"

	prom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ErrInvalidHelloRequest = errors.BadRequest("invalid hello request", "invalid_hello_request")
)

var (
	metricSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "server",
		Subsystem: "requests",
		Name:      "duration_sec",
		Help:      "server requests duration(sec).",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1},
	}, []string{"kind", "operation"})
)

type HelloRequest struct {
	Name string `json:"name"`
}

type HelloReply struct {
	Message string `json:"message"`
}

func SayHello(ctx http.Context) error {
	var req HelloRequest
	if err := ctx.BindVars(&req); err != nil {
		return ErrInvalidHelloRequest.WithCause(err)
	}
	http.SetOperation(ctx, "SayHello")

	h := ctx.Middleware(func(ctx context.Context, req any) (any, error) {
		return &HelloReply{
			Message: fmt.Sprintf("Hello %s!", req.(*HelloRequest).Name),
		}, nil
	})

	reply, err := h(ctx, &req)
	if err != nil {
		return err
	}

	return ctx.Result(stdhttp.StatusOK, reply)
}

func main() {
	logger := log.NewStdLogger(os.Stdout)
	log.SetLogger(logger)

	reg := prometheus.NewRegistry()
	reg.MustRegister(metricSeconds)

	httpSrv := http.NewServer(
		http.Address(":8000"),
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			metrics.Server(
				metrics.WithSeconds(prom.NewHistogram(metricSeconds)),
			),
		),
	)
	httpSrv.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	r := httpSrv.Route("/")
	r.GET("/hello/{name}", SayHello)

	app := kratos.New(
		kratos.Name("helloworld"),
		kratos.Server(httpSrv),
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
