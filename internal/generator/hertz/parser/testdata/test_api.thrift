namespace go testa.api

struct Request {
  1: required string name
  2: required string nick_name
}

struct Response {
  1: required string data
}

service TestAPI {
   Response testMethod(1: Request request)
}(api.service_group="test")