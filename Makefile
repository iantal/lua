.PHONY: protos

protos:
	protoc -I protos/ protos/lua.proto --go_out=plugins=grpc:protos
