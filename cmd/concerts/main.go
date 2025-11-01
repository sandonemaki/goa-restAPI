package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	goahttp "goa.design/goa/v3/http"

	// 生成されたパッケージにはgenプレフィックスを使用
	genconcerts "concerts/gen/concerts"
	genhttp "concerts/gen/http/concerts/server"
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

// mainはサービスをインスタンス化し、HTTPサーバーを起動します
func main() {
	// サービスのインスタンス化
	svc := &ConcertsService{}

	// 生成されたエンドポイントでラップ
	endpoints := genconcerts.NewEndpoints(svc)

	// HTTPハンドラーの構築
	mux := goahttp.NewMuxer()
	requestDecoder := goahttp.RequestDecoder
	responseEncoder := goahttp.ResponseEncoder
	handler := genhttp.New(endpoints, mux, requestDecoder, responseEncoder, nil, nil)

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
