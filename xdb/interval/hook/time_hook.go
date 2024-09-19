package hook

import "time"

type Time struct {
	Format string
}

func (t *Time) Input(row map[string]interface{}, fieldValue interface{}) (interface{}, error) {
	return fieldValue, nil
}

func (t *Time) Output(row map[string]interface{}, fieldValue interface{}) (interface{}, error) {
	if t.Format == "" {
		t.Format = "2006-01-02 15:04:05"
	}

	if value, ok := fieldValue.(*time.Time); ok {
		return value.Format(t.Format), nil
	}

	if value, ok := fieldValue.(time.Time); ok {
		return value.Format(t.Format), nil
	}

	return fieldValue, nil
}
