namespace go testa.api

include "test_api.thrift"

struct Response1 {
  1: required string data
}

// 测试服务
service TestAPI1 {
   Response1 testMethod(1: test_api.Request request)
}(api.service_group="test1" api.service_path="v1")