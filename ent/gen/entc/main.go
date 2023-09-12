package main

import (
	"log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	ex, err := entgql.NewExtension(
		entgql.WithWhereInputs(true), // GraphQLファイルにXxxWhereInput型を生成する
		entgql.WithSchemaGenerator(), // entからGraphQLの型定義を生成
	)
	if err != nil {
		log.Fatalf("creating entgql extension: %v", err)
	}

	opts := []entc.Option{
		// entc.TemplateDir("../ent/template"),
		entc.Extensions(ex),
		entc.FeatureNames(
			"sql/execquery",   // entのclientから生のSQLを実行できるようにする
			"schema/snapshot", // スキーマのスナップショットを作成して、コード生成が破綻しないようにする
			"sql/upsert",      // upsertを使えるようにする
			"privacy",
			"intercept",
		),
	}

	if err := entc.Generate("../ent/schema", &gen.Config{}, opts...); err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}

}
