package main

import (
	"fmt"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type HelloReply struct {
	Message string `json:"message"`
}

func SayHello(ctx http.Context) error {
	name := ctx.Vars().Get("name")

	reply := &HelloReply{
		Message: fmt.Sprintf("Hello %s!", name),
	}

	return ctx.JSON(200, reply)
}

func main() {
	httpSrv := http.NewServer(http.Address(":8000"))

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
