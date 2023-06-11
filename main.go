package main

import (
	"fmt"
	"github.com/jiangh156/mmo_game/apis"
	"github.com/jiangh156/mmo_game/core"
	"github.com/jiangh156/zinx/ziface"
	"github.com/jiangh156/zinx/znet"
)

// 当客户端建立连接的时候的hook函数
func OnConnectionAdd(conn ziface.IConnection) {
	//创建一个玩家
	player := core.NewPlayer(conn)
	//同步当前的PlayerID给客户端， 走MsgID:1 消息
	player.SyncPid()
	//同步当前玩家的初始化坐标信息给客户端，走MsgID:200消息
	player.BroadCastStartPosition()
	//将当前新上线玩家添加到worldManager中
	core.WorldMgrObj.AddPlayer(player)
	//将该连接绑定属性Pid
	conn.SetProperty("pid", player.Pid)

	//==============同步周边玩家上线信息，与现实周边玩家信息========
	player.SyncSurrounding()
	//=======================================================

	fmt.Println("=====> Player pidId = ", player.Pid, " arrived ====")
}

// 当当前连接断开之前的hook函数
func OnConnectionLost(conn ziface.IConnection) {
	pid, _ := conn.GetProperty("pid")
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))

	//触发玩家下线的任务
	player.Offline()
	fmt.Println("======> Player pid = ", pid, " offline... <=======")

}

func main() {
	//创建zinx server句柄
	s := znet.NewServer()

	//连接创建和销毁的Hook函数
	s.SetOnConnStart(OnConnectionAdd)
	s.SetOnConnStop(OnConnectionLost)

	//注册路由业务
	s.AddRouter(2, &apis.WorldChatApi{})
	s.AddRouter(3, &apis.MoveApi{})

	//启动服务
	s.Serve()
}
