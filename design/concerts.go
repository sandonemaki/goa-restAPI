package design

import (
	. "goa.design/goa/v3/dsl"
)

// サービス定義
var _ = Service("concerts", func() {
	Description("コンサートサービスは音楽コンサートのデータを管理します。")

	Method("list", func() {
		Description("オプションのページネーション付きで今後のコンサートを一覧表示します。")

		Payload(func() {
			Attribute("page", Int, "ページ番号", func() {
				Minimum(1)
				Default(1)
			})
			Attribute("limit", Int, "1ページあたりの項目数", func() {
				Minimum(1)
				Maximum(100)
				Default(10)
			})
		})

		Result(ArrayOf(Concert))

		HTTP(func() {
			GET("/concerts")

			// ページネーションのクエリパラメータ
			Param("page", Int, "ページ番号", func() {
				Minimum(1)
			})
			Param("limit", Int, "1ページあたりの項目数", func() {
				Minimum(1)
				Maximum(100)
			})

			Response(StatusOK) // Bodyを指定する必要はありません、Resultから推論されます
		})
	})

	Method("create", func() {
		Description("新しいコンサートエントリーを作成します。")

		Payload(ConcertPayload)
		Result(Concert)

		HTTP(func() {
			POST("/concerts")
			Response(StatusCreated)
		})
	})

	Method("show", func() {
		Description("IDで単一のコンサートを取得します。")

		Payload(func() {
			Attribute("concertID", String, "コンサートのUUID", func() {
				Format(FormatUUID)
			})
			Required("concertID")
		})

		Result(Concert)
		Error("not_found")

		HTTP(func() {
			GET("/concerts/{concertID}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("update", func() {
		Description("IDで既存のコンサートを更新します。")

		Payload(func() {
			Extend(ConcertPayload)
			Attribute("concertID", String, "更新するコンサートのID", func() {
				Format(FormatUUID)
			})
			Required("concertID")
		})

		Result(Concert, "更新されたコンサート")

		Error("not_found", ErrorResult, "コンサートが見つかりません")

		HTTP(func() {
			PUT("/concerts/{concertID}")

			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("delete", func() {
		Description("IDでシステムからコンサートを削除します。")

		Payload(func() {
			Attribute("concertID", String, "削除するコンサートのID", func() {
				Format(FormatUUID)
			})
			Required("concertID")
		})

		Error("not_found", ErrorResult, "コンサートが見つかりません")

		HTTP(func() {
			DELETE("/concerts/{concertID}")

			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
		})
	})
})

// データ型
var ConcertPayload = Type("ConcertPayload", func() {
	Description("コンサートの作成/更新に必要なデータ")

	Attribute("artist", String, "出演アーティスト/バンド", func() {
		MinLength(1)
		Example("The Beatles")
	})
	Attribute("date", String, "コンサート日付（YYYY-MM-DD）", func() {
		Pattern(`^\d{4}-\d{2}-\d{2}$`)
		Example("2024-01-01")
	})
	Attribute("venue", String, "コンサート会場", func() {
		MinLength(1)
		Example("The O2 Arena")
	})
	Attribute("price", Int, "チケット価格（USD）", func() {
		Minimum(1)
		Example(100)
	})
})

var Concert = Type("Concert", func() {
	Description("すべての詳細を含むコンサート")
	Extend(ConcertPayload)

	Attribute("id", String, "一意のコンサートID", func() {
		Format(FormatUUID)
	})
	Required("id", "artist", "date", "venue", "price")
})
