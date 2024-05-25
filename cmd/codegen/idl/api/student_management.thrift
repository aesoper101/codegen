namespace go api

struct QueryStudentRequest1 {
    1: string Num;
}

struct QueryStudentResponse1 {
    1: bool   Exist;
    2: string Num;
    3: string Name;
    4: string Gender;
    5: bool otherData; // 是否查询成功
}

struct InsertStudentRequest1 {
    1: string Num;
    2: string Name;
    3: string Gender;
}

struct InsertStudentResponse1 {
    1: bool Ok;
    2: string Msg;
}

service StudentManagement {
    QueryStudentResponse1 queryStudent1(1: QueryStudentRequest1 req)(api.get="student/query1");
    InsertStudentResponse1 insertStudent1(1: InsertStudentRequest1 req)(api.post="student/query2");
}(api.service_group="student")