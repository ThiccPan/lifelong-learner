# oapi-codegen --config=./api/openapi/apiconfig.yml ./api/openapi/users.yml

oapi-codegen -generate types,skip-prune  -o internal/users/openapi_types.gen.go  -package main ./api/openapi/users.yml
oapi-codegen -generate chi-server,skip-prune -o internal/users/openapi_server.gen.go -package main ./api/openapi/users.yml
oapi-codegen -generate spec,skip-prune   -o internal/users/openapi_spec.gen.go   -package main ./api/openapi/users.yml
# oapi-codegen -generate client,skip-prune -o openapi_client.gen.go -package api flamenco-openapi.yaml