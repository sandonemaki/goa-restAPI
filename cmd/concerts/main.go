package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"context"

	"github.com/google/uuid"
	goahttp "goa.design/goa/v3/http"

	// 生成されたパッケージにはgenプレフィックスを使用
	genconcerts "concerts/gen/concerts"
	genhttp "concerts/gen/http/concerts/server"

	"github.com/vmihailenco/msgpack/v5"
)

// ConcertsServiceはgenconcerts.Serviceインターフェースを実装
type ConcertsService struct {
	concerts []*genconcerts.Concert // インメモリストレージ
}

// オプションのページネーションで予定されているコンサートを一覧表示
func (m *ConcertsService) List(ctx context.Context, p *genconcerts.ListPayload) ([]*genconcerts.Concert, error) {
	start := (p.Page - 1) * p.Limit
	end := start + p.Limit
	if end > len(m.concerts) {
		end = len(m.concerts)
	}
	return m.concerts[start:end], nil
}

// 新しいコンサートエントリーを作成
func (m *ConcertsService) Create(ctx context.Context, p *genconcerts.ConcertPayload) (*genconcerts.Concert, error) {
	newConcert := &genconcerts.Concert{
		ID:     uuid.New().String(),
		Artist: *p.Artist,
		Date:   *p.Date,
		Venue:  *p.Venue,
		Price:  *p.Price,
	}
	m.concerts = append(m.concerts, newConcert)
	return newConcert, nil
}

// IDで単一のコンサートを取得
func (m *ConcertsService) Show(ctx context.Context, p *genconcerts.ShowPayload) (*genconcerts.Concert, error) {
	for _, concert := range m.concerts {
		if concert.ID == p.ConcertID {
			return concert, nil
		}
	}
	// 設計されたエラーを使用
	return nil, genconcerts.MakeNotFound(fmt.Errorf("concert not found: %s", p.ConcertID))
}

// IDで既存のコンサートを更新
func (m *ConcertsService) Update(ctx context.Context, p *genconcerts.UpdatePayload) (*genconcerts.Concert, error) {
	for i, concert := range m.concerts {
		if concert.ID == p.ConcertID {
			if p.Artist != nil {
				concert.Artist = *p.Artist
			}
			if p.Date != nil {
				concert.Date = *p.Date
			}
			if p.Venue != nil {
				concert.Venue = *p.Venue
			}
			if p.Price != nil {
				concert.Price = *p.Price
			}
			m.concerts[i] = concert
			return concert, nil
		}
	}
	return nil, genconcerts.MakeNotFound(fmt.Errorf("concert not found: %s", p.ConcertID))
}

// IDでシステムからコンサートを削除
func (m *ConcertsService) Delete(ctx context.Context, p *genconcerts.DeletePayload) error {
	for i, concert := range m.concerts {
		if concert.ID == p.ConcertID {
			m.concerts = append(m.concerts[:i], m.concerts[i+1:]...)
			return nil
		}
	}
	return genconcerts.MakeNotFound(fmt.Errorf("concert not found: %s", p.ConcertID))
}

type (
	// MessagePackエンコーダーの実装
	msgpackEnc struct {
		w http.ResponseWriter
	}

	// MessagePackデコーダーの実装
	msgpackDec struct {
		r *http.Request
	}
)

// カスタムエンコーダーコンストラクタ - MessagePackエンコーダーを作成
func msgpackEncoder(ctx context.Context, w http.ResponseWriter) goahttp.Encoder {
	return &msgpackEnc{w: w}
}

func (e *msgpackEnc) Encode(v any) error {
	e.w.Header().Set("Content-Type", "application/msgpack")
	return msgpack.NewEncoder(e.w).Encode(v)
}

// カスタムデコーダーコンストラクタ - 受信するMessagePackデータを処理
func msgpackDecoder(r *http.Request) goahttp.Decoder {
	return &msgpackDec{r: r}
}

func (d *msgpackDec) Decode(v any) error {
	return msgpack.NewDecoder(d.r.Body).Decode(v)
}

// mainはサービスをインスタンス化し、HTTPサーバーを起動します
func main() {
	// サービスのインスタンス化
	svc := &ConcertsService{}

	// 生成されたエンドポイントでラップ
	endpoints := genconcerts.NewEndpoints(svc)

	// HTTPハンドラーの構築
	mux := goahttp.NewMuxer()

	// クライアントの要望（Acceptヘッダー）に基づくスマートなエンコーダー選択
	encodeFunc := func(ctx context.Context, w http.ResponseWriter) goahttp.Encoder {
		accept := ctx.Value(goahttp.AcceptTypeKey).(string)

		// q値を含む複数のタイプを含むAcceptヘッダーを解析
		// 例：「application/json;q=0.9,application/msgpack」
		types := strings.Split(accept, ",")
		for _, t := range types {
			mt := strings.TrimSpace(strings.Split(t, ";")[0])
			switch mt {
			case "application/msgpack":
				return msgpackEncoder(ctx, w)
			case "application/json", "*/*":
				return goahttp.ResponseEncoder(ctx, w)
			}
		}

		// 迷ったときは、JSONが味方です！
		return goahttp.ResponseEncoder(ctx, w)
	}

	// クライアントが送信するもの（Content-Type）に基づくスマートなデコーダー選択
	decodeFunc := func(r *http.Request) goahttp.Decoder {
		if r.Header.Get("Content-Type") == "application/msgpack" {
			return msgpackDecoder(r)
		}
		return goahttp.RequestDecoder(r)
	}

	// カスタムエンコーダー/デコーダーを接続
	handler := genhttp.New(
		endpoints,
		mux,
		decodeFunc,
		encodeFunc,
		nil,
		nil,
	)

	// ハンドラーをmuxにマウント
	genhttp.Mount(mux, handler)

	// 新しいHTTPサーバーを作成
	port := "8080"
	server := &http.Server{Addr: ":" + port, Handler: mux}

	// サポートされているルートをログ出力
	for _, mount := range handler.Mounts {
		log.Printf("%q mounted on %s %s", mount.Method, mount.Verb, mount.Pattern)
	}

	// サーバーを起動（実行をブロックします）
	log.Printf("Starting concerts service on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
