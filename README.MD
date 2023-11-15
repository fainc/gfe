# GoFrame V2 Extension

拓展中间件: 


  | 类别       |       模块 |                    使用方法                    |                          说明                           |
  |:---------|---------:|:------------------------------------------:|:-----------------------------------------------------:|
  | 全局前置中间件  |     CORS |       middleware.CORSDefaultRegister       |                      默认的CORS跨域配置                      |
  | 全局前置中间件  |     Lang |        middleware.MultiLangRegister        |                  多语言，指定默认语言，从请求头读取语言                  |
  | 全局后置中间件  |   Logger |    middleware.Logger(options).Register     |           业务、错误日志（主要解决框架业务错误和系统错误没有区分开的问题）            |
  | 全局后置中间件  | Response | middleware.Response(defaultMime).Register  |        规范路由自适应数据输出，支持JSON、XML、HTML和Custom 自定义         |
  | 业务中间件    |      Jwt |      middleware.Jwt(jwtIns).Register       |           通过规范路由自动验证Token，支持声明免验、Redis吊销验证            |
  | 实例       |      Jwt |              ins.NewJwt(cgf)               | 支持 Validate / ParseRaw / Publish / IsRevoked / Revoke |
  | Helper   |  CtxUser |              helper.CtxUser()              |              支持 请求上下文 Get / Set    用户信息               |
  | Response | Response | response.StandError(ctx,msg,detail...) ... |                     内置错误封装等返回信息定义                     |

[//]: # (  | 全局前置中间件 |  Traffic |       middleware.Traffic&#40;&#41;.Regsiter        |                    接口速率和配额管理                      |)