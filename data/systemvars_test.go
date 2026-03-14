package data_test

import (
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/data"
)

func TestNewSystemVariables_Defaults(t *testing.T) {
	sv := data.NewSystemVariables()
	if sv == nil {
		t.Fatal("NewSystemVariables returned nil")
	}
	if sv.PageNumber != 1 {
		t.Errorf("PageNumber default = %d, want 1", sv.PageNumber)
	}
	if sv.TotalPages != 0 {
		t.Errorf("TotalPages default = %d, want 0", sv.TotalPages)
	}
	if sv.Row != 1 {
		t.Errorf("Row default = %d, want 1", sv.Row)
	}
	if sv.AbsRow != 1 {
		t.Errorf("AbsRow default = %d, want 1", sv.AbsRow)
	}
	if sv.Date.IsZero() {
		t.Error("Date should be set to current time")
	}
}

func TestSystemVariables_Get_PageNumber(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.PageNumber = 5
	v := sv.Get(data.SysVarPageNumber)
	if v != 5 {
		t.Errorf("Get(PageNumber) = %v, want 5", v)
	}
}

func TestSystemVariables_Get_TotalPages(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.TotalPages = 10
	if sv.Get(data.SysVarTotalPages) != 10 {
		t.Error("Get(TotalPages) should be 10")
	}
	// PageCount is an alias
	if sv.Get(data.SysVarPageCount) != 10 {
		t.Error("Get(PageCount) alias should be 10")
	}
}

func TestSystemVariables_Get_Date(t *testing.T) {
	sv := data.NewSystemVariables()
	d := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	sv.Date = d
	v := sv.Get(data.SysVarDate)
	if v != d {
		t.Errorf("Get(Date) = %v, want %v", v, d)
	}
}

func TestSystemVariables_Get_Row(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.Row = 7
	if sv.Get(data.SysVarRow) != 7 {
		t.Errorf("Get(Row#) = %v, want 7", sv.Get(data.SysVarRow))
	}
}

func TestSystemVariables_Get_AbsRow(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.AbsRow = 42
	if sv.Get(data.SysVarAbsRow) != 42 {
		t.Errorf("Get(AbsRow#) = %v, want 42", sv.Get(data.SysVarAbsRow))
	}
}

func TestSystemVariables_Get_Unknown(t *testing.T) {
	sv := data.NewSystemVariables()
	if sv.Get("NoSuchVar") != nil {
		t.Error("Get of unknown variable should return nil")
	}
}

func TestSystemVariables_Set_PageNumber(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.Set(data.SysVarPageNumber, 3)
	if sv.PageNumber != 3 {
		t.Errorf("PageNumber after Set = %d, want 3", sv.PageNumber)
	}
}

func TestSystemVariables_Set_TotalPages(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.Set(data.SysVarTotalPages, 20)
	if sv.TotalPages != 20 {
		t.Errorf("TotalPages after Set = %d, want 20", sv.TotalPages)
	}
}

func TestSystemVariables_Set_Date(t *testing.T) {
	sv := data.NewSystemVariables()
	d := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	sv.Set(data.SysVarDate, d)
	if sv.Date != d {
		t.Errorf("Date after Set = %v", sv.Date)
	}
}

func TestSystemVariables_Set_Row(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.Set(data.SysVarRow, 9)
	if sv.Row != 9 {
		t.Errorf("Row after Set = %d, want 9", sv.Row)
	}
}

func TestSystemVariables_Set_HierarchyLevel(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.Set(data.SysVarHierarchyLevel, 2)
	if sv.HierarchyLevel != 2 {
		t.Errorf("HierarchyLevel after Set = %d, want 2", sv.HierarchyLevel)
	}
}

func TestSystemVariables_Set_WrongType_NoOp(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.Set(data.SysVarPageNumber, "notanint") // wrong type → no-op
	if sv.PageNumber != 1 {
		t.Errorf("PageNumber should still be 1 after wrong-type Set, got %d", sv.PageNumber)
	}
}

func TestSystemVariables_ToParameters(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.PageNumber = 3
	sv.TotalPages = 10
	params := sv.ToParameters()
	if len(params) == 0 {
		t.Fatal("ToParameters returned empty slice")
	}
	found := false
	for _, p := range params {
		if p.Name == data.SysVarPageNumber && p.Value == 3 {
			found = true
		}
	}
	if !found {
		t.Error("ToParameters should include PageNumber=3")
	}
}
