gen-proto: PROTO_DIR = api
gen-proto: OUT_DIR = internal/proto

gen-proto:
	mkdir -p $(OUT_DIR)
	protoc \
		-I $(PROTO_DIR) \
		--go_out $(OUT_DIR) \
		--go-grpc_out $(OUT_DIR) \
		$(PROTO_DIR)/*.proto
