.PHONY: protos

protos:
	protoc -I protos/ protos/lua.proto --go_out=plugins=grpc:protos

result-protos:
	protoc -I protos/ protos/luaresult.proto --go_out=plugins=grpc:protos