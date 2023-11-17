package helper

import (
	"math"

	"github.com/gogf/gf/v2/database/gdb"
)

type model struct {
	m *gdb.Model
}

func Model(m *gdb.Model) *model {
	if m == nil {
		panic("model error")
	}
	return &model{m}
}

func (rec *model) CountWithPage(pageSize int) (rows int, page int, err error) {
	rows, err = rec.m.Count()
	if err != nil {
		return
	}
	if rows <= 0 || pageSize <= 0 {
		return rows, 1, nil
	}
	page = int(math.Ceil(float64(rows) / float64(pageSize)))
	return
}
