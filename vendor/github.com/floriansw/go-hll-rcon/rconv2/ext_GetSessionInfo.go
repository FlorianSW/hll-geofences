package rconv2

type Vector2D struct {
	X float64
	Y float64
}

type mapData struct {
	SectorSize float64
	// by default the center of the map is at vector 0,0, however, some maps (like Carentan Skirmish)
	// move the center of the map (as visual to the player) on the x and/or y-axis. MapCenterOffset is the Vector
	// that describes this offset. It has 0,0 by default.
	MapCenterOffset Vector2D
}

// maps describe the default configuration of a map, or a specific different config if a map
// is different to the default.
var maps = map[GetSessionInfoGameMode]map[string]mapData{
	GetSessionInfoGameModeSkirmish: {
		"default": {
			SectorSize: 13926,
		},
		"CARENTAN": {
			SectorSize:      13926,
			MapCenterOffset: Vector2D{X: 150, Y: -110},
		},
		"MORTAIN": {
			SectorSize:      13926,
			MapCenterOffset: Vector2D{X: 100, Y: 0},
		},
		"ST MARIE DU MONT": {
			SectorSize:      13926,
			MapCenterOffset: Vector2D{X: 0, Y: -27852.799},
		},
		"DRIEL": {
			SectorSize:      13926,
			MapCenterOffset: Vector2D{X: -20, Y: 28190},
		},
		"EL ALAMEIN": {
			SectorSize:      13926,
			MapCenterOffset: Vector2D{X: -7500, Y: 0},
		},
		"STALINGRAD": {
			SectorSize:      13926,
			MapCenterOffset: Vector2D{X: 150, Y: -110},
		},
		"JUNO BEACH": {
			SectorSize: 13926,
		},
	},
	GetSessionInfoGameModeWarfare: {
		// all older maps (SME, SMDM, etc) have a default sector width and height
		"default": {
			SectorSize: 19840,
		},
		// Carentan has a slightly higher sector size
		"CARENTAN": {
			SectorSize: 20160,
		},
		// newer maps have a 200x200m grid schema
		"ELSENBORN RIDGE": {
			SectorSize: 20000,
		},
		"MORTAIN": {
			SectorSize: 20000,
		},
		"TOBRUK": {
			SectorSize: 20000,
		},
		"SMOLENSK": {
			SectorSize: 20000,
		},
		"JUNO BEACH": {
			SectorSize: 20000,
		},
	},
}

func (m GetSessionInfoResponse) mapData() *mapData {
	gm := m.GameMode
	if gm == GetSessionInfoGameModeOffensive {
		// for what mapData is concerned, Offensive behaves (so far) the same as warfare.
		gm = GetSessionInfoGameModeWarfare
	}
	if mode, exists := maps[m.GameMode]; !exists {
		return nil
	} else if md, exists := mode[m.MapName]; exists {
		return &md
	} else if d, exists := mode["default"]; exists {
		return &d
	} else {
		return nil
	}
}

// GridSize returns the size in meters of a grid square on the map, depending on the current game mode.
func (m GetSessionInfoResponse) GridSize() float64 {
	if d := m.mapData(); d != nil {
		return d.SectorSize
	}
	return 0
}
