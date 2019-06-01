
package apis

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"Project03/mmo_game_server/core"
	"Project03/mmo_game_server/pb"
	"Project03/zinx/net"
	"Project03/zinx/ziface"
)

//世界聊天 路由业务
type WorldChat struct {
	net.BaseRouter
}

func (wc *WorldChat) Handle(request ziface.IRequest) {
	//1 解析客户端传递进来的protobuf数据
	proto_msg := &pb.Talk{}
	if err := proto.Unmarshal(request.GetMsg().GetMsgData(), proto_msg);err != nil {
		fmt.Println("Talk message unmarshal error ", err)
		return
	}

	//通过获取链接属性，得到当前的玩家ID
	pid, err := request.GetConnection().GetProperty("pid")
	if err != nil {
		fmt.Println("get Pid error ", err)
		return
	}

	//通过pid 来得到对应的player对象
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))

	// 当前的聊天数据广播给全部的在线玩家
	//当前玩家的windows客户端发送过来的消息
	player.SendTalkMsgToAll(proto_msg.GetContent())

}
