package core

import "fmt"

// 定义一些AOI的边界值
const (
	AOI_MIN_X  int = 85
	AOI_MAX_X  int = 410
	AOI_CNTS_X int = 10
	AOI_MIN_Y  int = 75
	AOI_MAX_Y  int = 400
	AOI_CNTS_Y int = 20
)

/*
AOI区域管理模块
*/
type AOIManager struct {
	MinX  int           //区域左边界坐标
	MaxX  int           //区域右边界坐标
	CntsX int           //x方向格子的数量
	MinY  int           //区域上边界坐标
	MaxY  int           //区域下边界坐标
	CntsY int           //y方向的格子数量
	grids map[int]*Grid //当前区域中都有哪些格子，key=格子ID， value=格子对象
}

/*
初始化一个AOI区域
*/
func NewAOIManager(minX, maxX, cntsX, minY, maxY, cntsY int) *AOIManager {
	aoiMgr := &AOIManager{
		MinX:  minX,
		MaxX:  maxX,
		CntsX: cntsX,
		MinY:  minY,
		MaxY:  maxY,
		CntsY: cntsY,
		grids: make(map[int]*Grid),
	}

	//给AOI初始化区域中所有的格子
	for y := 0; y < cntsY; y++ {
		for x := 0; x < cntsX; x++ {
			//计算格子ID
			//格子编号：id = idy *nx + idx  (利用格子坐标得到格子编号)
			gid := y*cntsX + x

			//初始化一个格子放在AOI中的map里，key是当前格子的ID
			aoiMgr.grids[gid] = NewGrid(gid,
				aoiMgr.MinX+x*aoiMgr.gridWidth(),
				aoiMgr.MinX+(x+1)*aoiMgr.gridWidth(),
				aoiMgr.MinY+y*aoiMgr.gridLength(),
				aoiMgr.MinY+(y+1)*aoiMgr.gridLength())
		}
	}

	return aoiMgr
}

// 得到每个格子在x轴方向的宽度
func (m *AOIManager) gridWidth() int {
	return (m.MaxX - m.MinX) / m.CntsX
}

// 得到每个格子在x轴方向的长度
func (m *AOIManager) gridLength() int {
	return (m.MaxY - m.MinY) / m.CntsY
}

// 打印信息方法
func (m *AOIManager) String() string {
	s := fmt.Sprintf("AOIManagr:\nminX:%d, maxX:%d, cntsX:%d, minY:%d, maxY:%d, cntsY:%d\n Grids in AOI Manager:\n",
		m.MinX, m.MaxX, m.CntsX, m.MinY, m.MaxY, m.CntsY)
	for _, grid := range m.grids {
		s += fmt.Sprintln(grid)
	}
	return s
}

// 根据格子GID得到周边九宫格格子集合
func (m *AOIManager) GetSurroundGridsByGid(gID int) (grids []*Grid) {
	//判断gID是否在AOIManager中
	if _, ok := m.grids[gID]; !ok {
		return nil
	}

	//初始化grids返回值切片，将当前gid本身假如九宫格切片中
	grids = append(grids, m.grids[gID])

	//需要gID的左边是否有格子？右边是否有格子
	//需要通过gID得到当前格子x轴的编号---idx = gID % nx;
	idx := gID % m.CntsX

	//判断idx编号是否左边是否还有格子
	if idx > 0 {
		grids = append(grids, m.grids[gID-1])
	}

	//判断idx编号是否右边是否还有格子
	if idx < m.CntsX-1 {
		grids = append(grids, m.grids[gID+1])
	}
	//将x轴当前的格子都取出，进行遍历，分别得到每个格子上下是否还有格子
	gidsX := make([]int, 0, len(grids))
	for _, v := range grids {
		gidsX = append(gidsX, v.GID)
	}

	//在遍历gidsX集合中每个格子的gid上下是否还有格子
	for _, v := range gidsX {
		idy := v / m.CntsX
		// gid上边是否还有格子
		if idy > 0 {
			grids = append(grids, m.grids[v-m.CntsX])
		}
		if idy < m.CntsY-1 {
			grids = append(grids, m.grids[v+m.CntsX])
		}
	}
	return
}

// 通过x,y横纵坐标得到当前的GID格子编号
func (m *AOIManager) GetGidByPos(x, y float32) int {
	idx := (int(x) - m.MinX) / m.gridWidth()
	idy := (int(y) - m.MinY) / m.gridLength()
	return idy*m.CntsX + idx
}

// 通过x,y横纵坐标得到周边九宫格内全部的PlayerIDs
func (m *AOIManager) GetPidsByPos(x, y float32) (playerIDs []int) {
	//得到当前玩家的GID格子id
	gID := m.GetGidByPos(x, y)

	//通过GID得到周边九宫格信息
	grids := m.GetSurroundGridsByGid(gID)

	//将九宫格的信息里的全部的Player
	for _, v := range grids {
		playerIDs = append(playerIDs, v.GetPlayerIDs()...)
		//fmt.Printf("====> grid ID: %d, pids: %v ====", v.GID, v.playerIDs)
	}
	return
}

// 添加一个PlayerID到一个格子中
func (m *AOIManager) AddPidToGrid(pId, gID int) {
	m.grids[gID].Add(pId)
}

// 移除一个格子中的PlayerID
func (m *AOIManager) RemovePidFromGrid(pId, gID int) {
	m.grids[gID].Remove(pId)
}

// 通过GID获取全部的PlayerID
func (m *AOIManager) GetPidsBuGid(gID int) (playerIDs []int) {
	return m.grids[gID].GetPlayerIDs()
}

// 通过坐标将Player添加到一个格子中
func (m *AOIManager) AddToGridByPos(pId int, x, y float32) {
	gID := m.GetGidByPos(x, y)
	m.grids[gID].Add(pId)

}

// 通过坐标把一个Player从一个格子中删除
func (m *AOIManager) RemoveFromGridByPos(pID int, x, y float32) {
	gID := m.GetGidByPos(x, y)
	m.grids[gID].Remove(pID)
}
