gen-proto: PROTO_DIR = api
gen-proto: OUT_DIR = internal/proto

gen-proto:
	buf generate
