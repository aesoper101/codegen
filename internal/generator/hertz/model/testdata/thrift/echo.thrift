namespace go api

struct EchoResponse {
    1: string message;
}

struct EchoRequest {
    1: string message;
}

service Echo {
    EchoResponse echo(1: EchoRequest request);
}

