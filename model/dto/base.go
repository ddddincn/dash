package dto

import (
	"dash/model/param"
	"dash/utils"
	"math"
	"reflect"
)

type BaseDTO struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
type Page struct {
	Contents    interface{} `json:"contents"`
	Pages       int         `json:"pages"`
	Total       int64       `json:"total"`
	RPP         int         `json:"rpp"`
	PageNum     int         `json:"page_num"`
	HasNext     bool        `json:"has_next"`
	HasPrevious bool        `json:"has_previous"`
	IsFirst     bool        `json:"is_first"`
	IsLast      bool        `json:"is_last"`
	IsEmpty     bool        `json:"is_empty"`
	HasContent  bool        `json:"has_content"`
}

func NewPage(contents interface{}, totalCount int64, page param.Page) *Page {
	var contentsLen int
	r := reflect.ValueOf(contents)

	if !r.IsNil() && r.Kind() != reflect.Slice {
		panic("not slice")
	} else {
		contentsLen = r.Len()
	}
	totalPage := utils.IfElse(page.PageSize == 0, 1, int(math.Ceil(float64(totalCount)/float64(page.PageSize)))).(int)
	dtoPage := &Page{
		Contents:    contents,
		Total:       totalCount,
		Pages:       totalPage,
		PageNum:     page.PageNum,
		RPP:         page.PageNum,
		HasNext:     page.PageNum+1 < totalPage,
		HasPrevious: page.PageNum > 0,
		IsFirst:     page.PageNum == 0,
		IsLast:      page.PageNum+1 == totalPage,
		IsEmpty:     contentsLen == 0,
		HasContent:  contentsLen > 0,
	}
	return dtoPage
}
