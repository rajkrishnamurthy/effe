// Code generated by Effe. DO NOT EDIT.

//+build !effeinject

package main

import (
	"fmt"
	"gopkg.in/h2non/gentleman.v2"
)

func C(service CService) CFunc {
	return func() (*gentleman.Response, error) {
		err := service.Step1()
		if err != nil {
			return nil, err
		}
		responsePtrVal, err := func() (*gentleman.Response, error) {
			cli := gentleman.New()
			cli.URI("http://example.com")
			req := cli.Request()
			req.Method(POST)
			return cli.Send()
		}()
		if err != nil {
			return responsePtrVal, err
		}
		return responsePtrVal, nil
	}
}
func NewCImpl() *CImpl {
	return &CImpl{step1FieldFunc: step1()}
}

type CService interface {
	Step1() error
}
type CImpl struct {
	step1FieldFunc func() error
}
type CFunc func() (*gentleman.Response,

	error)

func (c *CImpl) Step1() error { return c.step1FieldFunc() }
