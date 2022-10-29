package testgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/vincoll/vigie/foundation/web"
	"github.com/vincoll/vigie/pkg/business/core/probe"
	v0Web "github.com/vincoll/vigie/pkg/business/web"
)

// Handlers manages the set of product endpoints.
type Handlers struct {
	Test probe.Core
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var nvt probe.VigieTest
	if err := web.Decode(r, &nvt); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	vt, err := h.Test.Create(ctx, nvt, v.Now)
	if err != nil {
		if errors.Is(err, probe.ErrNotFoundProbe) {
			return v0Web.NewRequestError(err, http.StatusConflict)
		}
		return fmt.Errorf("user[%+v]: %w", &vt, err)
	}

	return web.Respond(ctx, w, nil, http.StatusCreated)

}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	return web.Respond(ctx, w, nil, http.StatusNoContent)

}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	return web.Respond(ctx, w, nil, http.StatusNoContent)

}

func (h *Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	return web.Respond(ctx, w, nil, http.StatusOK)

}

func (h *Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	return web.Respond(ctx, w, nil, http.StatusOK)

}

func (h Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is my testgtrp"))
}
