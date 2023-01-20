package testgrp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vincoll/vigie/foundation/tools"
	"github.com/vincoll/vigie/foundation/web"
	"github.com/vincoll/vigie/pkg/business/core/probe"
)

// Handlers manages the set of product endpoints.
type Handlers struct {
	Test *probe.Core
}

func (h Handlers) Create(c *gin.Context) {

	ctx := c.Request.Context()
	v, err := web.GetValues(ctx)
	if err != nil {
		//return web.NewShutdownError("web value missing from context")
	}

	var nvt probe.VigieTest
	if err := web.Decode(c.Request, &nvt); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "unable to decode payload",
			"error":   err.Error(),
			"payload": c.Request.Body,
		},
		)
		return
	}

	err = h.Test.Create(ctx, &nvt, v.Now)
	if err != nil {
		if errors.Is(err, probe.ErrNotFoundProbe) {
			//	return v0Web.NewRequestError(err, http.StatusConflict)
		}
		//	return fmt.Errorf("user[%+v]: %w", &vt, err)
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"message": "implement Name ?"})
}

func (h Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	return web.Respond(ctx, w, nil, http.StatusNoContent)

}

func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	return web.Respond(ctx, w, nil, http.StatusNoContent)

}

func (h Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	return web.Respond(ctx, w, nil, http.StatusOK)

}

func (h Handlers) QueryByID(c *gin.Context) {

	ctx := c.Request.Context()

	/*
		v, err := web.GetValues(ctx)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{
				"message": "web value missing from context",
				"error":   err.Error(),
			})
			return
		}
	*/

	id := c.Param("id")
	if !tools.IsValidUUID(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "ID have a wrong format",
			"error":   id + "is not a UUID",
		},
		)
		return
	}

	vt, err := h.Test.GetByID(ctx, id, time.Now())
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"message": "Unable to get the test.",
			"error":   err.Error(),
		},
		)
		return
	}
	/*
		_, err := json.Marshal(vt)
		if err != nil {
			fmt.Println(err)
			return
		}
	*/
	c.IndentedJSON(http.StatusOK, vt)
	return
}

func (h Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is my testgtrp"))
}
