package testgrp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vincoll/vigie/foundation/tools"
	"github.com/vincoll/vigie/foundation/web"
	"github.com/vincoll/vigie/pkg/business/core/probemgmt"
)

// Handlers manages the set of product endpoints.
type Handlers struct {
	Test *probemgmt.Core
}

func (h Handlers) Create(c *gin.Context) {

	ctx := c.Request.Context()
	/*
		v, _ := web.GetValues(ctx)

			if err != nil {
				//return web.NewShutdownError("web value missing from context")
			}
	*/
	var vtr VigieTestREST
	if err := web.Decode(c.Request, &vtr); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "unable to decode payload",
			"error":   err.Error(),
			"payload": c.Request.Body,
		},
		)
		return
	}

	// Write it to DB
	err := h.Test.Create(ctx, vtr.toVigieTest())
	if err != nil {
		if errors.Is(err, probemgmt.ErrDBUnavailable) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to write to DB", "reason": "DB unavailable", "status": http.StatusInternalServerError})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to write to DB", "reason": err.Error(), "status": http.StatusInternalServerError})
		}
		return
	}

	metaResp := web.VigieReponseMetaAPI{
		HTTPStatus: http.StatusCreated,
		Message:    "Test created (%d)",
	}
	c.IndentedJSON(metaResp.HTTPStatus, web.NewVigieResponseAPI(nil, metaResp))
	return
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

		metaResp := web.VigieReponseMetaAPI{
			HTTPStatus: http.StatusBadRequest,
			ErrorType:  "ID have a wrong format",
			ErrorTrace: "...",
			Message:    id + "is not a UUID",
		}
		c.IndentedJSON(metaResp.HTTPStatus, web.NewVigieResponseAPI(nil, metaResp))
		return
	}

	vt, err := h.Test.GetByID(ctx, id, time.Now())
	if err != nil {

		err2 := errors.Unwrap(err)

		if err2.Error() == probemgmt.ErrDBNotFound.Error() {

			metaResp := web.VigieReponseMetaAPI{
				HTTPStatus: http.StatusInternalServerError,
				ErrorType:  "Test does not exists",
				ErrorTrace: "...",
				Message:    err.Error(),
			}
			c.IndentedJSON(metaResp.HTTPStatus, web.NewVigieResponseAPI(nil, metaResp))
			return
		}

		metaResp := web.VigieReponseMetaAPI{
			HTTPStatus: http.StatusInternalServerError,
			ErrorType:  "Unable to get the test",
			ErrorTrace: "...",
			Message:    err.Error(),
		}
		c.IndentedJSON(metaResp.HTTPStatus, web.NewVigieResponseAPI(nil, metaResp))
		return
	}
	vtj, _ := vt.ToVigieTestJSON()
	// https://stackoverflow.com/questions/49611868/best-way-to-handle-interfaces-in-http-response

	metaResp := web.VigieReponseMetaAPI{
		HTTPStatus: http.StatusOK,
		Message:    "QueryByType",
	}

	c.IndentedJSON(http.StatusOK, web.NewVigieResponseAPI(vtj, metaResp))
	return
}

func (h Handlers) QueryByType(c *gin.Context) {

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

	typ := c.Param("type")

	// Temp, list of valid types should be centralized somewhere else
	validTypes := [2]string{"icmp", "tcp"}
	if tools.StringInSlice(typ, validTypes[:]) == false {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid",
			"error":   typ + "is not a Valid Type",
		},
		)
		return
	}

	vts, err := h.Test.GetByType(ctx, typ, time.Now())
	vtjs := make([]probemgmt.VigieTestJSON, 0, len(vts))

	for _, vt := range vts {
		vtj, _ := vt.ToVigieTestJSON()
		vtjs = append(vtjs, *vtj)
	}

	if err != nil {

		err2 := errors.Unwrap(err)

		if err2.Error() == probemgmt.ErrDBNotFound.Error() {
			c.IndentedJSON(http.StatusNotFound, gin.H{
				"message": "Test does not exists",
				"error":   err.Error(),
			})
			return
		}

		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"message": "Unable to get the test.",
			"error":   err.Error(),
		},
		)
		return
	}

	metaResp := web.VigieReponseMetaAPI{
		HTTPStatus: http.StatusOK,
		Message:    "QueryByType",
	}

	c.IndentedJSON(http.StatusOK, web.NewVigieResponseAPI(vts, metaResp))
	return
}

func (h Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is my testgtrp"))
}
