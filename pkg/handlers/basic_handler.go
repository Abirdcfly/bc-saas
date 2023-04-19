/*
Copyright 2023 The Bestchains Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handler

import (
	"encoding/base64"

	"github.com/bestchains/bc-saas/pkg/contracts"
	"github.com/bestchains/bc-saas/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

type KeyValue struct {
	Index   string `json:"index,omitempty"`
	KID     string `json:"kid,omitempty"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message,omitempty"`
}

type VerifyStatus struct {
	Status bool   `json:"status"`
	Reason string `json:"reason"`
}

type BasicHandler struct {
	basic *contracts.Basic
}

func NewBasicHandler(basic *contracts.Basic) BasicHandler {
	return BasicHandler{
		basic: basic,
	}
}

// Total
func (h *BasicHandler) Total(ctx *fiber.Ctx) error {
	total, err := h.basic.Total()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(map[string]interface{}{
		"total": total,
	})
}

func (h *BasicHandler) PutValue(ctx *fiber.Ctx) error {
	var err error

	kv := new(KeyValue)
	err = ctx.BodyParser(kv)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// validate message
	rawMsg, err := base64.StdEncoding.DecodeString(kv.Message)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errors.Wrap(err, "invalid message").Error())
	}
	message := new(utils.Message)
	if err = message.Unmarshal(rawMsg); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errors.Wrap(err, "invalid message").Error())
	}

	kid, err := h.basic.PutValue(message, kv.Value)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(&KeyValue{
		KID: kid,
	})

}

func (h *BasicHandler) VerifyValue(ctx *fiber.Ctx) error {
	var err error

	kvArgs := new(KeyValue)
	err = ctx.BodyParser(kvArgs)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	kv, err := h.getValue(*kvArgs)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if kv.Value != kvArgs.Value {
		return ctx.JSON(&VerifyStatus{
			Status: false,
			Reason: "value mismatch",
		})
	}

	return ctx.JSON(&VerifyStatus{
		Status: true,
		Reason: "value match",
	})
}

func (h *BasicHandler) GetValue(ctx *fiber.Ctx) error {
	arg := KeyValue{
		Index: ctx.Query("index"),
		KID:   ctx.Query("kid"),
	}

	if arg.Index == "" && arg.KID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "must provide depository index or kid")
	}

	kv, err := h.getValue(arg)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(kv)
}

func (h *BasicHandler) getValue(kv KeyValue) (*KeyValue, error) {
	if kv.Index != "" {
		value, err := h.basic.GetValueByIndex(kv.Index)
		if err != nil {
			return nil, err
		}
		kv.Value = value
	} else if kv.KID != "" {
		value, err := h.basic.GetValueByKID(kv.KID)
		if err != nil {
			return nil, err
		}
		kv.Value = value
	}

	return &kv, nil
}