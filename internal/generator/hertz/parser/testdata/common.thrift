namespace go common

include "base.thrift"

struct BaseRequest {
   1: string id;
}

struct BaseResponse {
   1: string id;
}

service CommonService extends base.BaseService {
    BaseResponse CommonPing(1: BaseRequest request) (api.get = '/commonping');
}