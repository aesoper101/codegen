namespace go hello

include "common.thrift"

struct HelloRequest {
    1: string name
}


service HelloService {
    common.BaseResponse welcome(1: HelloRequest message)
    string welcome1(1: string message)
}