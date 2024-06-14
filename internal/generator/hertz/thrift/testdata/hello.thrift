namespace go backend.api


include "common.thrift"

struct HelloRequest {
    1: string message
    2: common.BaseRequest baseRequest
}

service HelloEveryOneService {
    common.BaseResponse welcome(1: HelloRequest request) (api.get = "/welcome")
}

service HelloService extends HelloEveryOneService{
    common.BaseResponse hello(1: HelloRequest request) (api.get = "/hello")
    common.BaseResponse hello2(1: string baseRequest) (api.get = "/hello2")
}