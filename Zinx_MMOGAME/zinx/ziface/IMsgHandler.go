
package ziface

/*
 抽象的消息管理模块  存放router集合的
 */

 type IMsgHandler interface {
 	//添加路由到map集合中
 	AddRouter(msgID uint32, router IRouter)
 	//调度路由， 根据MsgID
 	DoMsgHandler(request IRequest)
 	//启动Worker工作池
 	StartWorkerPool()
 	//将消息添加到Worker工作池中 （将消息发送给对应的消息队列）
 	SendMsgToTaskQueue(request IRequest)
 }
