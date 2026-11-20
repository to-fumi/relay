// Package api は api/openapi.yaml から生成した HTTP の契約（型とサーバーインターフェース）。
package api

//go:generate go tool -modfile=../../tools/go.mod oapi-codegen -config oapi-codegen.yaml ../../../api/openapi.yaml
