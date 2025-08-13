package handler

// import (
// 	"dash/consts"
// 	"dash/service"
// 	"dash/service/assembler"
// 	"dash/utils"

// 	"github.com/gin-gonic/gin"
// )

// type SheetHandler struct {
// 	PostService   service.PostService
// 	PostAssembler assembler.PostAssembler
// }

// func NewSheetHandler(postService service.PostService, postAssembler assembler.PostAssembler) *SheetHandler {
// 	return &SheetHandler{
// 		PostService:   postService,
// 		PostAssembler: postAssembler,
// 	}
// }

// func (s *SheetHandler) Sheet(ctx *gin.Context) (interface{}, error) {
// 	slug, err := utils.ParamString(ctx, "slug")
// 	if err != nil {
// 		return nil, err
// 	}
// 	post, err := s.PostService.GetPostBySlug(ctx, slug)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if post.Type != consts.PostTypeSheet {
// 		return nil, nil
// 	}
// 	postVO, err := s.PostAssembler.ConvertToDetailVO(ctx, post)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return postVO, nil
// }
