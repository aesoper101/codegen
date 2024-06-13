namespace go backend.api


include "common.thrift"

struct HelloRequest {
    1: string message
    2: common.BaseRequest baseRequest
}

service HelloService extends common.CommonService{
    common.BaseResponse hello(1: HelloRequest request) (api.get = "/hello")
    common.BaseResponse hello2(1: string baseRequest) (api.get = "/hello2")
}