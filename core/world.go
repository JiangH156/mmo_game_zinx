package core

import "sync"

// 当前游戏的世界总管理模块
type WorldManager struct {
	//AOIManager 当前世界地图AOI的管理模块
	AoiMgr *AOIManager
	//当前全部在线的Players集合
	Players map[int32]*Player
	//保护Player集合的锁
	pLock sync.RWMutex
}

// 提供一个对外的世界管理模块的句柄（全局）
var WorldMgrObj *WorldManager

// 初始化方法
func init() {
	WorldMgrObj = &WorldManager{
		//创建世界AOI地图规划
		AoiMgr: NewAOIManager(AOI_MIN_X, AOI_MAX_X, AOI_CNTS_X, AOI_MIN_Y, AOI_MAX_Y, AOI_CNTS_Y),
		//初始化Players集合
		Players: make(map[int32]*Player),
	}
}

// 添加一个玩家
func (w *WorldManager) AddPlayer(player *Player) {
	w.pLock.Lock()
	w.Players[player.Pid] = player
	w.pLock.Unlock()

	//将Player添加到AOIManager中
	w.AoiMgr.AddToGridByPos(int(player.Pid), player.X, player.Z)
}

// 删除一个玩家
func (w *WorldManager) RemovePlayer(pid int32) {
	//得到当前玩家
	player := w.Players[pid]
	//将玩家从AOIManager中删除
	w.AoiMgr.RemoveFromGridByPos(int(pid), player.X, player.Z)

	//将玩家从世界管理中删除
	w.pLock.Lock()
	delete(w.Players, pid)
	w.pLock.Unlock()

}

// 通过玩家ID查询Player对象
func (w *WorldManager) GetPlayerByPid(pid int32) *Player {
	w.pLock.RLock()
	defer w.pLock.RUnlock()
	return w.Players[pid]
}

// 获取全部的在线玩家
func (w *WorldManager) GetAllPlayers() []*Player {
	w.pLock.RLock()
	defer w.pLock.RUnlock()

	players := make([]*Player, 0)
	for _, v := range w.Players {
		players = append(players, v)
	}
	return players
}
