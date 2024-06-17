namespace go api

include "common.thrift"
include "hello.thrift"

struct Request {
	1: string message
}

struct Response {
	1: string message
}

service Echo extends hello.HelloService{
    Response echo(1: Request req)
    common.BaseResponse noRequest(1: common.BaseRequest message)
}

service Echo2 extends hello.HelloService{
    Response echo(1: Request req)
    common.BaseResponse noRequest(1: common.BaseRequest message)
}