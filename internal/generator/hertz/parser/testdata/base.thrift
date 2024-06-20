namespace go api.common

struct BaseRequest {
   1: string id;
}

struct BaseResponse {
   1: string id;
}

service BaseService {
    BaseResponse Ping(1: BaseRequest request) (api.get = '/ping');
}