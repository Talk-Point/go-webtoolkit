package repository

import (
	"context"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/Talk-Point/go-webtoolkit/pkg/v2/errors"
	"github.com/Talk-Point/go-webtoolkit/pkg/v2/query"
)

type Entity interface {
	DocId() string
	SetDocId(id string)
	UniqFields() map[string]interface{}
}

type Repository[T Entity, TT any] interface {
	Get(ctx context.Context, opts *query.QueryOptions) (*PaginationResult[T], error)
	GetByID(ctx context.Context, id string) (*T, error)
	Create(ctx context.Context, obj T) (*string, error)
	CreateEasy(ctx context.Context, obj T) (*string, error)
	Update(ctx context.Context, id string, data map[string]interface{}) error
	Delete(ctx context.Context, id string) error
}

type repository[T Entity, TT any] struct {
	Db         *firestore.Client
	Collection string
	Ressource  string
}

func NewFirebaseRepository[T Entity, TT any](db *firestore.Client, ressoucre string) Repository[T, TT] {
	collectionName := strings.ToLower(ressoucre) + "s"

	return &repository[T, TT]{
		Db:         db,
		Collection: collectionName,
		Ressource:  ressoucre,
	}
}

type PaginationResult[T any] struct {
	Items   []T             `json:"items"`
	Limit   int             `json:"limit"`
	Next    string          `json:"next"`
	Prev    string          `json:"prev"`
	Filters *[]query.Filter `json:"filters,omitempty"`
}

func (r *repository[T, TT]) Get(ctx context.Context, opts *query.QueryOptions) (*PaginationResult[T], error) {
	if opts == nil {
		opts = &query.QueryOptions{
			Limit: 100,
		}
	}
	q := r.Db.Collection(r.Collection).Limit(opts.Limit)

	isFirstPage := true
	if opts.Next != "" {
		isFirstPage = false
		q = q.StartAfter(opts.Next)
	}
	if opts.Previous != "" {
		isFirstPage = false
		q = q.EndBefore(opts.Previous)
	}

	if opts.OrderBy != "" {
		direction := firestore.Asc
		if opts.OrderByDirection == query.Desc {
			direction = firestore.Desc
		}
		if opts.OrderBy == "id" {
			q = q.OrderBy(firestore.DocumentID, direction)
		} else {
			q = q.OrderBy(opts.OrderBy, direction)
		}
	}

	for _, f := range opts.Filters {
		if f.Field == "id" {
			docRef := r.Db.Collection(r.Collection).Doc(f.Value.(string))
			q = q.Where(firestore.DocumentID, "==", docRef)
		} else {
			q = q.Where(f.Field, f.Operator.ToFireStoreOperator(), f.Value)
		}
	}

	page := q.Documents(ctx)
	docs, err := page.GetAll()
	if err != nil {
		return nil, err
	}

	var nextPageKey string
	if len(docs) >= opts.Limit && len(docs) > 0 {
		nextPageKey = docs[len(docs)-1].Ref.ID
	} else {
		nextPageKey = ""
	}

	var prevPageKey string
	if len(docs) > 0 {
		if !isFirstPage {
			prevPageKey = docs[0].Ref.ID
		}
	} else {
		prevPageKey = ""
	}

	objs := make([]T, 0)
	for _, doc := range docs {
		obj := new(T)
		if err := (*doc).DataTo(obj); err != nil {
			return nil, err
		}
		(*obj).SetDocId((*doc).Ref.ID)
		objs = append(objs, *obj)
	}

	return &PaginationResult[T]{
		Items:   objs,
		Limit:   opts.Limit,
		Next:    nextPageKey,
		Prev:    prevPageKey,
		Filters: &opts.Filters,
	}, nil
}

func (r *repository[T, TT]) GetByID(ctx context.Context, id string) (*T, error) {
	doc, err := r.Db.Collection(r.Collection).Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	obj := new(T)
	if err := (*doc).DataTo(obj); err != nil {
		return nil, err
	}
	(*obj).SetDocId((*doc).Ref.ID)

	return obj, nil
}

func (r *repository[T, TT]) Create(ctx context.Context, obj T) (*string, error) {
	var docID string
	err := r.Db.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		query := r.Db.Collection(r.Collection).Query
		for field, value := range obj.UniqFields() {
			query = query.Where(field, "==", value)
		}
		documents, err := query.Documents(ctx).GetAll()
		if err != nil {
			return err
		}
		if len(documents) > 0 {
			return &errors.ErrorAlreadyExists{
				ErrorDetail: errors.ErrorDetail{
					Resource: r.Ressource,
					Field:    "Reference",
					Value:    "",
					Message:  r.Ressource + " with Reference already exists",
				},
			}
		}

		val := reflect.ValueOf(obj).Elem()
		if val.Kind() == reflect.Struct {
			for _, fieldName := range []string{"CreatedAt", "UpdatedAt"} {
				field := val.FieldByName(fieldName)
				if field.IsValid() && field.CanSet() && field.Type() == reflect.TypeOf(time.Now()) {
					field.Set(reflect.ValueOf(time.Now()))
				}
			}
		}

		docRef := r.Db.Collection(r.Collection).NewDoc()
		tx.Set(docRef, obj)
		docID = docRef.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &docID, nil
}

func (r *repository[T, TT]) CreateEasy(ctx context.Context, obj T) (*string, error) {
	var docID string
	val := reflect.ValueOf(obj).Elem()
	if val.Kind() == reflect.Struct {
		for _, fieldName := range []string{"CreatedAt", "UpdatedAt"} {
			field := val.FieldByName(fieldName)
			if field.IsValid() && field.CanSet() && field.Type() == reflect.TypeOf(time.Now()) {
				field.Set(reflect.ValueOf(time.Now()))
			}
		}
	}
	docRef := r.Db.Collection(r.Collection).NewDoc()
	_, err := docRef.Set(ctx, obj)
	if err != nil {
		return nil, err
	}
	docID = docRef.ID

	return &docID, nil
}

func (r *repository[T, TT]) Update(ctx context.Context, id string, data map[string]interface{}) error {
	updates := []firestore.Update{}
	for k, v := range data {
		updates = append(updates, firestore.Update{
			Path:  k,
			Value: v,
		})
	}
	_, err := r.Db.Collection(r.Collection).Doc(id).Update(ctx, updates)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository[T, TT]) Delete(ctx context.Context, id string) error {
	_, err := r.Db.Collection(r.Collection).Doc(id).Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}
