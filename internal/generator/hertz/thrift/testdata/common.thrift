namespace go common

struct BaseRequest {
   1: string id;
}

struct BaseResponse {
   1: string id;
}

service CommonService {
    BaseResponse Ping(1: BaseRequest request) (api.get = '/ping');
}