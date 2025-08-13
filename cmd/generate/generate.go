package main

import (
	"dash/config"
	"dash/dal"
	"dash/log"

	"gorm.io/gen"
)

func main() {
	conf := config.NewConfig()
	logger := log.NewLogger(conf)
	gormLogger := log.NewGormLogger(conf, logger)
	DB := dal.NewGormDB(conf, gormLogger)
	g := gen.NewGenerator(gen.Config{
		Mode:              gen.WithDefaultQuery,
		OutPath:           "./dal",
		ModelPkgPath:      "./model/entity",
		FieldNullable:     true,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
	})

	g.UseDB(DB)

	g.ApplyBasic(
		g.GenerateModel("category", gen.FieldType("type", "consts.CategoryType")),
		g.GenerateModel("menu"),
		g.GenerateModel("option", gen.FieldType("type", "consts.OptionType")),
		g.GenerateModel("post", gen.FieldType("type", "consts.PostType"), gen.FieldType("status", "consts.PostStatus"), gen.FieldType("editor_type", "consts.EditorType")),
		g.GenerateModel("post_category"),
		g.GenerateModel("post_tag"),
		g.GenerateModel("tag"),
		g.GenerateModel("theme_setting"),
		g.GenerateModel("user", gen.FieldType("mfa_type", "consts.MFAType")),
	)
	g.Execute()
}
