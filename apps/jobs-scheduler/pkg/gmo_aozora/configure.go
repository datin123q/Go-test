package gmo_aozora

import (
	corporateclient "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	webhookclient "github.com/gmoaozora/gmo-aozora-api-go/webhookclient"
)

type GMO_ENV = string

var (
	STG_ENV  GMO_ENV = "stg"
	PROD_ENV GMO_ENV = "prod"
)

func NewStgConfigure() *corporateclient.Configuration {
	cfg := &corporateclient.Configuration{
		BasePath:      "https://stg-api.gmo-aozora.com/ganb/api/corporation/v1",
		DefaultHeader: make(map[string]string),
		UserAgent:     "Swagger-Codegen/1.0.0/go",
	}
	return cfg
}

func NewProdConfigure() *corporateclient.Configuration {
	cfg := &corporateclient.Configuration{
		BasePath:      "https://api.gmo-aozora.com/ganb/api/corporation/v1",
		DefaultHeader: make(map[string]string),
		UserAgent:     "Swagger-Codegen/1.0.0/go",
	}
	return cfg
}

func NewClientConfigure(e GMO_ENV) *corporateclient.Configuration {
	switch e {
	case STG_ENV:
		return NewStgConfigure()
	case PROD_ENV:
		return NewProdConfigure()
	default:
		return NewStgConfigure()
	}
}

func NewStgWeHookConfigure() *webhookclient.Configuration {
	cfg := &webhookclient.Configuration{
		BasePath:      "https://stg-api.gmo-aozora.com/ganb/api/webhooks/v1",
		DefaultHeader: make(map[string]string),
		UserAgent:     "Swagger-Codegen/1.0.0/go",
	}
	return cfg
}

func NewProdWeHookConfigure() *webhookclient.Configuration {
	cfg := &webhookclient.Configuration{
		BasePath:      "https://api.gmo-aozora.com/ganb/api/webhooks/v1",
		DefaultHeader: make(map[string]string),
		UserAgent:     "Swagger-Codegen/1.0.0/go",
	}
	return cfg
}

func NewWebHookConfigure(e GMO_ENV) *webhookclient.Configuration {
	switch e {
	case STG_ENV:
		return NewStgWeHookConfigure()
	case PROD_ENV:
		return NewProdWeHookConfigure()
	default:
		return NewStgWeHookConfigure()
	}
}
