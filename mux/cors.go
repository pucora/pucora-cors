package mux

import (
	"bufio"
	"io"
	"log"

	pucoracors "github.com/pucora/pucora-cors/v2"
	"github.com/pucora/lura/v2/config"
	"github.com/pucora/lura/v2/logging"
	"github.com/pucora/lura/v2/router/mux"
	"github.com/rs/cors"
)

// New returns a mux.HandlerMiddleware (which implements the http.Handler interface)
// with the CORS configuration defined in the ExtraConfig.
func New(e config.ExtraConfig) mux.HandlerMiddleware {
	return NewWithLogger(e, nil)
}

func NewWithLogger(e config.ExtraConfig, l logging.Logger) mux.HandlerMiddleware {
	tmp := pucoracors.ConfigGetter(e)
	if tmp == nil {
		return nil
	}
	cfg, ok := tmp.(pucoracors.Config)
	if !ok {
		return nil
	}

	if len(cfg.AllowOrigins) == 0 {
		cfg.AllowOrigins = []string{"*"}
	}
	if len(cfg.AllowHeaders) == 0 {
		cfg.AllowHeaders = []string{"*"}
	}

	c := cors.New(cors.Options{
		AllowedOrigins:       cfg.AllowOrigins,
		AllowedMethods:       cfg.AllowMethods,
		AllowedHeaders:       cfg.AllowHeaders,
		ExposedHeaders:       cfg.ExposeHeaders,
		AllowCredentials:     cfg.AllowCredentials,
		AllowPrivateNetwork:  cfg.AllowPrivateNetwork,
		OptionsPassthrough:   cfg.OptionsPassthrough,
		OptionsSuccessStatus: cfg.OptionsSuccessStatus,
		Debug:                cfg.Debug,
		MaxAge:               int(cfg.MaxAge.Seconds()),
	})
	if l == nil || !cfg.Debug {
		return c
	}

	r, w := io.Pipe()
	c.Log = log.New(w, "", log.LstdFlags)
	go writeLog(r, l)

	return c
}

func writeLog(r *io.PipeReader, l logging.Logger) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		l.Debug("[CORS]", scanner.Text())
	}
}
