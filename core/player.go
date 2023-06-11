package core

import (
	"fmt"
	"github.com/jiangh156/mmo_game/pb"
	"github.com/jiangh156/zinx/ziface"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"sync"
)

// 玩家对象
type Player struct {
	Pid  int32              //玩家ID
	Conn ziface.IConnection //当前玩家的连接（用于和客户端的连接）
	X    float32            //平面X坐标
	Y    float32            //高度
	Z    float32            //平面的Y坐标
	V    float32            //旋转角度（0-360）
}

// Player ID 生成器
var PidGen int32 = 1  //用来生成玩家ID的计数器
var IDLock sync.Mutex //保护PidGen的Mutex

// 创建一个玩家的方法
func NewPlayer(conn ziface.IConnection) *Player {
	//生成一个玩家ID
	IDLock.Lock()
	id := PidGen
	PidGen++
	IDLock.Unlock()

	//创建一个玩家对象
	p := &Player{
		Pid:  id,
		Conn: conn,
		X:    float32(160 + rand.Intn(10)), //随机在160坐标点，基于X轴若干偏移
		Y:    0,
		Z:    float32(140 + rand.Intn(20)), //随机在140坐标点，基于Y轴若干偏移
		V:    0,                            //角度为0
	}
	return p
}

/*
提供一个发送给客户端消息的方法
主要是将pb的protobuf数据序列化之后，再调用zinx的SendMsg方法
*/
func (p *Player) SendMsg(msgId uint32, data proto.Message) {
	//将proto Message结构体序列化，转换成二进制
	msg, err := proto.Marshal(data)
	if err != nil {
		fmt.Println("marshal err:", err)
		return
	}

	//将二进制文件，通过zinx框架的SendMsg将数据发送给客户端
	if p.Conn == nil {
		fmt.Println("connection in player is nil")
		return
	}
	if err = p.Conn.SendMsg(msgId, msg); err != nil {
		fmt.Println("Player SendMsg error!")
		return
	}
}

// 告知客户端玩家Pid，同步已经生成的玩家ID给客户端
func (p *Player) SyncPid() {
	//组建MsgID：1的proto数据
	proto_msg := &pb.SyncPid{
		Pid: p.Pid,
	}
	//将消息发送给客户端
	p.SendMsg(1, proto_msg)
}

// 广播玩家组件的出生地点
func (p *Player) BroadCastStartPosition() {
	//组建MsgID：200的proto数据
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2, //Tp2代表广播的位置坐标
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//将消息发送给客户端
	p.SendMsg(200, proto_msg)
}

// 玩家广播世界聊天消息
func (p *Player) Talk(content string) {
	//1 组件MsgID：200 proto数据
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  1, //代表聊天广播
		Data: &pb.BroadCast_Content{
			Content: content,
		},
	}

	//2 得到当前世界所有的在线玩家
	players := WorldMgrObj.GetAllPlayers()

	//3 向所有的玩家发送数据（包括自己）
	for _, player := range players {
		// player分别给对应的客户端发送消息
		player.SendMsg(200, proto_msg)
	}
}

// 给当前玩家周边的(九宫格内)玩家广播自己的位置，让他们显示自己
func (p *Player) SyncSurrounding() {
	//1 获取当前玩家周围的玩家有那些（九宫格）
	pids := WorldMgrObj.AoiMgr.GetPidsByPos(p.X, p.Z)
	players := make([]*Player, 0, len(pids))
	for _, pid := range pids {
		players = append(players, WorldMgrObj.GetPlayerByPid(int32(pid)))
	}
	//2 将当前的位置信息通过MsgId：200 发给周围的玩家（让其他玩家看到自己）
	//2.1 组建msgId：200 proto数据
	msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2, //代表广播坐标
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}
	//2.2 周围的玩家都向各自的客户端发送200的消息，proto_msg
	for _, player := range players {
		player.SendMsg(200, msg)
	}
	//3 将周围的全部玩家的位置信息发送给当前的玩家MsgID：202 客户端（当自己看到其他玩家）
	//3.1 组建msgId：202 proto数据
	//3.1.1 制作pb.Player 切片
	playersData := make([]*pb.Player, 0, len(players))
	for _, player := range players {
		p := &pb.Player{
			Pid: player.Pid,
			P: &pb.Position{
				X: player.X,
				Y: player.Y,
				Z: player.Z,
				V: player.V,
			},
		}
		playersData = append(playersData, p)
	}
	//3.1.2 封装SyncPlayer Proto数据
	SyncPlayersMsg := &pb.SyncPlayers{
		Ps: playersData[:],
	}

	//3.2 将组建号的数据发送给当前玩家的客户端
	p.SendMsg(202, SyncPlayersMsg)
}

// 广播并更新当前玩家的坐标
func (p *Player) UpdatePos(x float32, y float32, z float32, v float32) {
	//更新当前玩家player对象的坐标
	p.X = x
	p.Y = y
	p.Z = z
	p.V = v
	//组建广播proto协议 MsgID；200 Tp4
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  4,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: x,
				Y: y,
				Z: z,
				V: v,
			},
		},
	}
	players := p.GetSurroundingPlayers()
	//依次给每个玩家对应的客户端发送当前玩家位置更新的信息
	for _, player := range players {
		player.SendMsg(200, proto_msg)
	}
}

// 获取当前玩家的周边玩家AOI九宫格之内的玩家
func (p *Player) GetSurroundingPlayers() []*Player {
	//得到当前AOI九宫格内的所有玩家PID
	pids := WorldMgrObj.AoiMgr.GetPidsByPos(p.X, p.Z)
	//将所有的pid对应的Player放到Players中
	players := make([]*Player, 0, len(pids))
	for _, pid := range pids {
		players = append(players, WorldMgrObj.GetPlayerByPid(int32(pid)))
	}
	return players
}

// 玩家下线业务
func (p *Player) Offline() {
	//得到当前玩家周边的九宫格内的玩家
	players := p.GetSurroundingPlayers()

	//给周围玩家广播MsgId：201消息
	proto_msg := &pb.SyncPid{
		Pid: p.Pid,
	}
	for _, player := range players {
		player.SendMsg(201, proto_msg)
	}
	WorldMgrObj.AoiMgr.RemoveFromGridByPos(int(p.Pid), p.X, p.Z)
	WorldMgrObj.RemovePlayer(p.Pid)
}
