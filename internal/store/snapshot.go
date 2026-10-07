package store

import (
	"sort"
	"time"

	"example.com/apiary-colony-service/internal/domain"
)

// Snapshot 是一次一致性读取的结果，供聚合与校验使用。
type Snapshot struct {
	Apiaries          map[string]domain.Apiary
	Hives             map[string]domain.Hive
	InspectionsByHive map[string][]domain.Inspection
	QueensByHive      map[string][]domain.Queen
	DueAtByHive       map[string]time.Time
	GeneratedAt       time.Time
}

// InspectionDue 描述一只蜂箱下一次应当安排巡检的时间。
type InspectionDue struct {
	HiveID string
	DueAt  time.Time
}

// InspectionDues 返回蜂场下已经登记过巡检安排的蜂箱。
func (s Snapshot) InspectionDues(apiaryID string) []InspectionDue {
	hives := s.ApiaryHives(apiaryID)
	dues := make([]InspectionDue, 0, len(hives))
	for _, hive := range hives {
		due, ok := s.DueAtByHive[hive.ID]
		if !ok {
			continue
		}
		dues = append(dues, InspectionDue{HiveID: hive.ID, DueAt: due})
	}
	return dues
}

// ApiaryHives 返回蜂场下的蜂箱，按编号升序稳定排序。
func (s Snapshot) ApiaryHives(apiaryID string) []domain.Hive {
	hives := make([]domain.Hive, 0, len(s.Hives))
	for _, hive := range s.Hives {
		if hive.ApiaryID == apiaryID {
			hives = append(hives, hive)
		}
	}
	sort.Slice(hives, func(i, j int) bool { return hives[i].ID < hives[j].ID })
	return hives
}

// LatestInspection 返回蜂箱最近一次巡检；观察时间相同时按编号取较大者。
func (s Snapshot) LatestInspection(hiveID string) (domain.Inspection, bool) {
	items := s.InspectionsByHive[hiveID]
	if len(items) == 0 {
		return domain.Inspection{}, false
	}
	latest := items[0]
	for _, item := range items[1:] {
		if item.ObservedAt.After(latest.ObservedAt) {
			latest = item
			continue
		}
		if item.ObservedAt.Equal(latest.ObservedAt) && item.ID > latest.ID {
			latest = item
		}
	}
	return latest, true
}

// LayingQueen 返回蜂箱当前处于产卵状态的蜂王。
func (s Snapshot) LayingQueen(hiveID string) (domain.Queen, bool) {
	for _, item := range s.QueensByHive[hiveID] {
		if item.IsLaying() {
			return item, true
		}
	}
	return domain.Queen{}, false
}

func (s Snapshot) Queens(hiveID string) []domain.Queen {
	items := s.QueensByHive[hiveID]
	out := make([]domain.Queen, len(items))
	copy(out, items)
	return out
}
