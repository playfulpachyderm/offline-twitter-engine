package tracing

import (
	"context"
	"time"

	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

type SpanID uint64

type Span struct {
	ID   SpanID `db:"rowid" json:"id"`
	Name string `db:"name" json:"name"`

	StartTime persistence.Timestamp `db:"start_time" json:"start_time"`
	EndTime   persistence.Timestamp `db:"end_time" json:"end_time"`
	IsActive  bool

	Children []*Span
	ParentID SpanID `db:"parent_id" json:"parent_id"`
}

func (s *Span) get_active_span() *Span {
	if len(s.Children) == 0 {
		return s
	}
	last_child := s.Children[len(s.Children)-1]
	if last_child.IsActive {
		return last_child.get_active_span()
	}
	return s
}

func (s *Span) End() {
	s.EndTime = persistence.Timestamp{time.Now()}
	s.IsActive = false
}

func (s Span) Duration() time.Duration {
	return s.EndTime.Time.Sub(s.StartTime.Time)
}

func (s *Span) AddChild(name string) *Span {
	ret := &Span{Name: name, StartTime: persistence.Timestamp{time.Now()}, IsActive: true, Children: []*Span{}}
	s.Children = append(s.Children, ret)
	return ret
}

type key string

const TRACE_KEY = key("TRACE")

func InitTrace(ctx context.Context, name string) (context.Context, *Span) {
	top_level_span := &Span{Name: name, StartTime: persistence.Timestamp{time.Now()}, IsActive: true, Children: []*Span{}}
	new_ctx := context.WithValue(ctx, TRACE_KEY, top_level_span)
	return new_ctx, top_level_span
}

func GetActiveSpan(ctx context.Context) *Span {
	span, is_ok := ctx.Value(TRACE_KEY).(*Span)
	if !is_ok {
		panic(ctx.Value(TRACE_KEY))
	}
	return span.get_active_span()
}

// Database
// --------

func (db *DB) SaveSpan(s *Span) {
	if s.ID == 0 {
		// Do create
		result, err := db.DB.NamedExec(`
			insert into spans (name, start_time, end_time, parent_id) values (:name, :start_time, :end_time, nullif(:parent_id, 0))
		`, s)
		if err != nil {
			panic(err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			panic(err)
		}
		s.ID = SpanID(id)

		// Save children recursively
		for _, child := range s.Children {
			child.ParentID = s.ID
			db.SaveSpan(child)
		}
	} else {
		// Do update
		result, err := db.DB.NamedExec(`
			update spans
	           set end_time = :end_time
	         where rowid = :rowid
		`, s)
		if err != nil {
			panic(err)
		}
		count, err := result.RowsAffected()
		if err != nil {
			panic(err)
		}
		if count != 1 {
			panic(fmt.Errorf("Got span with ID (%d), so attempted update, but it doesn't exist", s.ID))
		}
	}
}

func (db *DB) GetSpanByID(id SpanID) (ret Span, err error) {
	err = db.DB.Get(&ret, `
		select rowid, name, cals, carbs, protein, fat, sugar, alcohol, water, potassium, calcium, sodium,
			   magnesium, phosphorus, iron, zinc, mass, price, density, cook_ratio
		  from spans
		 where rowid = ?
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Span{}, ErrNotInDB
	}
	return
}
