package testgrp

import (
	"context"
	"errors"
	"fmt"
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
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("Test created (%d)", vtr.Metadata.UID), "status": http.StatusCreated})
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
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "ID have a wrong format",
			"error":   id + "is not a UUID",
		},
		)
		return
	}

	vt, err := h.Test.GetByID(ctx, id, time.Now())
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

	validTypes := [2]string{"icmp", "tcp"}
	if tools.StringInSlice(typ, validTypes[:]) == false {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid",
			"error":   typ + "is not a Valid Type",
		},
		)
		return
	}

	vt, err := h.Test.GetByType(ctx, typ, time.Now())
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
