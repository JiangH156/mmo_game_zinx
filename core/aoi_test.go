package core

import (
	"fmt"
	"testing"
)

func TestNewAOIManager(t *testing.T) {
	//初始化AOIManager
	aoiMgr := NewAOIManager(100, 300, 4, 200, 450, 5)

	fmt.Println(aoiMgr)
}

func TestGetSurroundGridsByGid(t *testing.T) {
	//初始化AOIManager
	aoiMgr := NewAOIManager(0, 250, 5, 0, 250, 5)

	for gid, _ := range aoiMgr.grids {
		//得到当前gid的周边九宫格信息
		grids := aoiMgr.GetSurroundGridsByGid(gid)
		fmt.Println("gid: ", gid, "grids len", len(grids))
		gIDs := make([]int, 0, len(grids))
		for _, grid := range grids {
			gIDs = append(gIDs, grid.GID)
		}
		fmt.Println("surrounding grid IDs are ", gIDs)
	}
}
