package main

import (
	"log/slog"

	"github.com/jackvonhouse/sticky-corner-bypass/winapi"
)

func main() {
	if err := winapi.New().Proccess(); err != nil {
		slog.Error(err.Error())
	}
}
