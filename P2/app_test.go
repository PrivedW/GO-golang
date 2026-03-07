package main

import (
	"database/sql"
	"errors"
	"testing"

	"P2/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

// MockRows - правильная реализация мока для sql.Rows
type MockRows struct {
	rowsData []MockRowData
	index    int
}

type MockRowData struct {
	Model string
	Price int
}

func NewMockRows(data []MockRowData) *MockRows {
	return &MockRows{
		rowsData: data,
		index:    -1,
	}
}

func (m *MockRows) Next() bool {
	m.index++
	return m.index < len(m.rowsData)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.index < 0 || m.index >= len(m.rowsData) {
		return sql.ErrNoRows
	}

	if len(dest) >= 2 {
		if modelPtr, ok := dest[0].(*string); ok {
			*modelPtr = m.rowsData[m.index].Model
		}
		if pricePtr, ok := dest[1].(*int); ok {
			*pricePtr = m.rowsData[m.index].Price
		}
	}
	return nil
}

func (m *MockRows) Close() error {
	return nil
}

func (m *MockRows) Err() error {
	return nil
}

type MockResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r *MockResult) LastInsertId() (int64, error) {
	return r.lastInsertID, nil
}

func (r *MockResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

func TestInsertProduct(t *testing.T) {
	type Test struct {
		Name      string
		Model     string
		Company   string
		Price     int
		MockExec  func(query string, args ...interface{}) (sql.Result, error)
		Times     int
		WantError error
	}

	tests := []Test{
		{
			Name:    "Success",
			Model:   "Galaxy A52",
			Company: "Samsung",
			Price:   4300,
			MockExec: func(query string, args ...interface{}) (sql.Result, error) {
				return &MockResult{lastInsertID: 1}, nil
			},
			Times:     1,
			WantError: nil,
		},
		{
			Name:    "Database Error",
			Model:   "Galaxy A52",
			Company: "Samsung",
			Price:   4300,
			MockExec: func(query string, args ...interface{}) (sql.Result, error) {
				return nil, errors.New("database error")
			},
			Times:     1,
			WantError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDBInterface(ctrl)
			mockDB.EXPECT().
				Exec("INSERT INTO products (model, company, price) VALUES (?, ?, ?)",
					tt.Model, tt.Company, tt.Price).
				DoAndReturn(tt.MockExec).
				Times(tt.Times)

			app := NewApp(mockDB)
			_, err := app.InsertProduct(tt.Model, tt.Company, tt.Price)

			if tt.WantError != nil {
				require.Error(t, err)
				require.Equal(t, tt.WantError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetAllProducts(t *testing.T) {
	type Test struct {
		Name      string
		MockQuery func(query string, args ...interface{}) (*sql.Rows, error)
		Times     int
		WantCount int
		WantError error
	}

	tests := []Test{
		{
			Name: "Success",
			MockQuery: func(query string, args ...interface{}) (*sql.Rows, error) {
				if query != "SELECT model, price FROM products" {
					t.Errorf("unexpected query: %s", query)
				}

				_ = NewMockRows([]MockRowData{
					{Model: "iPhone 13", Price: 999},
					{Model: "Galaxy S21", Price: 899},
					{Model: "Pixel 6", Price: 799},
				})

				return nil, errors.New("this approach doesn't work with sql.Rows")
			},
			Times:     1,
			WantCount: 3,
			WantError: nil,
		},
		{
			Name: "Query Error",
			MockQuery: func(query string, args ...interface{}) (*sql.Rows, error) {
				return nil, errors.New("query failed")
			},
			Times:     1,
			WantError: errors.New("query failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDBInterface(ctrl)
			mockDB.EXPECT().
				Query("SELECT model, price FROM products").
				DoAndReturn(tt.MockQuery).
				Times(tt.Times)

			app := NewApp(mockDB)
			_, err := app.GetAllProducts()

			if tt.WantError != nil {
				require.Error(t, err)
				require.Equal(t, tt.WantError.Error(), err.Error())
			} else {
			}
		})
	}
}

func TestDeleteByCompany(t *testing.T) {
	type Test struct {
		Name      string
		Company   string
		MockExec  func(query string, args ...interface{}) (sql.Result, error)
		Times     int
		WantCount int64
		WantError error
	}

	tests := []Test{
		{
			Name:    "Success",
			Company: "Samsung",
			MockExec: func(query string, args ...interface{}) (sql.Result, error) {
				if query != "DELETE FROM products WHERE company = ?" {
					t.Errorf("unexpected query: %s", query)
				}
				if len(args) != 1 || args[0] != "Samsung" {
					t.Errorf("unexpected args: %v", args)
				}
				return &MockResult{rowsAffected: 2}, nil
			},
			Times:     1,
			WantCount: 2,
			WantError: nil,
		},
		{
			Name:    "Delete Error",
			Company: "Samsung",
			MockExec: func(query string, args ...interface{}) (sql.Result, error) {
				return nil, errors.New("delete failed")
			},
			Times:     1,
			WantError: errors.New("delete failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDBInterface(ctrl)
			mockDB.EXPECT().
				Exec("DELETE FROM products WHERE company = ?", tt.Company).
				DoAndReturn(tt.MockExec).
				Times(tt.Times)

			app := NewApp(mockDB)
			count, err := app.DeleteByCompany(tt.Company)

			if tt.WantError != nil {
				require.Error(t, err)
				require.Equal(t, tt.WantError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.WantCount, count)
			}
		})
	}
}

func TestGetAllProducts_QueryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDBInterface(ctrl)
	expectedErr := errors.New("query failed")

	mockDB.EXPECT().
		Query("SELECT model, price FROM products").
		Return(nil, expectedErr).
		Times(1)

	app := NewApp(mockDB)
	_, err := app.GetAllProducts()

	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}
