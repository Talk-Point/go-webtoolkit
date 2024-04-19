package query

import (
	"testing"
)

func TestQueryOptions(t *testing.T) {
	url := "/users?limit=98&next=123&prev=124&salechannel=webshop&sort=-createdAt"

	q, err := NewQueryOptionsFromUrlString(url)
	if err != nil {
		t.Fatal(err)
	}

	if q.Limit != 98 {
		t.Errorf("Expected 98, got %d", q.Limit)
	}
	if q.Next != "123" {
		t.Errorf("Expected 123, got %s", q.Next)
	}
	if q.Previous != "124" {
		t.Errorf("Expected 124, got %s", q.Previous)
	}
	if q.OrderBy != "createdAt" {
		t.Errorf("Expected createdAt, got %s", q.OrderBy)
	}
	if q.OrderByDirection != Desc {
		t.Errorf("Expected Desc, got %s", q.OrderByDirection)
	}
}

func TestQueryOptionsOrderBy(t *testing.T) {
	t.Run("TestQueryOptionsOrderBy ASC", func(t *testing.T) {
		url := "/users?sort=createdAt"

		q, err := NewQueryOptionsFromUrlString(url)
		if err != nil {
			t.Fatal(err)
		}

		if q.OrderBy != "createdAt" {
			t.Errorf("Expected createdAt, got %s", q.OrderBy)
		}
		if q.OrderByDirection != Asc {
			t.Errorf("Expected Asc, got %s", q.OrderByDirection)
		}
	})

	t.Run("TestQueryOptionsOrderBy DESC", func(t *testing.T) {
		url := "/users?sort=-createdAt"

		q, err := NewQueryOptionsFromUrlString(url)
		if err != nil {
			t.Fatal(err)
		}

		if q.OrderBy != "createdAt" {
			t.Errorf("Expected createdAt, got %s", q.OrderBy)
		}
		if q.OrderByDirection != Desc {
			t.Errorf("Expected Desc, got %s", q.OrderByDirection)
		}
	})

	t.Run("TestQueryOptionsOrderByDefault", func(t *testing.T) {
		url := "/users"

		q, err := NewQueryOptionsFromUrlString(url)
		if err != nil {
			t.Fatal(err)
		}

		if q.OrderBy != "id" {
			t.Errorf("Expected id, got %s", q.OrderBy)
		}
		if q.OrderByDirection != Desc {
			t.Errorf("Expected Desc, got %s", q.OrderByDirection)
		}
	})
}

func TestQueryOptionsLimit(t *testing.T) {
	t.Run("TestQueryOptionsLimit", func(t *testing.T) {
		url := "/users?limit=98&ps=123&salechannel=webshop&sort=-createdAt"

		q, err := NewQueryOptionsFromUrlString(url)
		if err != nil {
			t.Fatal(err)
		}

		if q.Limit != 98 {
			t.Errorf("Expected 98, got %d", q.Limit)
		}
	})

	t.Run("TestQueryOptionsLimitDefault", func(t *testing.T) {
		url := "/users"

		q, err := NewQueryOptionsFromUrlString(url)
		if err != nil {
			t.Fatal(err)
		}

		if q.Limit != 30 {
			t.Errorf("Expected 30, got %d", q.Limit)
		}
	})

	t.Run("TestQueryOptionsLimitMax", func(t *testing.T) {
		url := "/users?limit=1000"

		q, err := NewQueryOptionsFromUrlString(url)
		if err != nil {
			t.Fatal(err)
		}

		if q.Limit != 100 {
			t.Errorf("Expected 100, got %d", q.Limit)
		}
	})
}

func TestFilter(t *testing.T) {
	t.Run("TestFilter easy", func(t *testing.T) {

		url := "/users?limit=98&next=123&prev=124&salechannel=webshop&sort=-createdAt"

		q, err := NewFilterFromUrlString(url)
		if err != nil {
			t.Fatal(err)
		}

		if len(q) != 1 {
			t.Errorf("Expected 1, got %d", len(q))
		}
		if q[0].Field != "salechannel" {
			t.Errorf("Expected salechannel, got %s", q[0].Field)
		}
		if q[0].Value != "webshop" {
			t.Errorf("Expected webshop, got %s", q[0].Value)
		}
		if q[0].Operator != Eq {
			t.Errorf("Expected eq, got %s", q[0].Operator)
		}
	})

	t.Run("TestFilter with __", func(t *testing.T) {
		url := "/users?limit=98&next=123&prev=124&salechannel__gt=webshop&sort=-createdAt"

		q, err := NewFilterFromUrlString(url)
		if err != nil {
			t.Fatal(err)
		}

		if len(q) != 1 {
			t.Errorf("Expected 1, got %d", len(q))
		}
		if q[0].Field != "salechannel" {
			t.Errorf("Expected salechannel, got %s", q[0].Field)
		}
		if q[0].Value != "webshop" {
			t.Errorf("Expected webshop, got %s", q[0].Value)
		}
		if q[0].Operator != Gt {
			t.Errorf("Expected gt, got %s", q[0].Operator)
		}
	})
}

func TestQueryFormValues(t *testing.T) {
	t.Run("TestQueryFormValues", func(t *testing.T) {
		url := "/users?limit=98&next=123&prev=124&salechannel=webshop&sort=-createdAt"

		q, err := NewQueryOptionsFromUrlString(url)
		if err != nil {
			t.Fatal(err)
		}

		if q.Limit != 98 {
			t.Errorf("Expected 98, got %d", q.Limit)
		}
		if q.Next != "123" {
			t.Errorf("Expected 123, got %s", q.Next)
		}
		if q.Previous != "124" {
			t.Errorf("Expected 124, got %s", q.Previous)
		}
		if q.OrderBy != "createdAt" {
			t.Errorf("Expected createdAt, got %s", q.OrderBy)
		}
		if q.OrderByDirection != Desc {
			t.Errorf("Expected Desc, got %s", q.OrderByDirection)
		}
	})
}
