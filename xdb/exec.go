package xdb

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// 一般用Prepared Statements和Exec()完成INSERT, UPDATE, DELETE操作
func exec(db *sql.DB, _sql string, args ...any) (res sql.Result, err error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	stmt, err := tx.Prepare(_sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	res, err = stmt.Exec(args...)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func execTx(tx *sql.Tx, _sql string, args ...any) (res sql.Result, err error) {
	stmt, err := tx.Prepare(_sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	res, err = stmt.Exec(args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func query(db *sql.DB, _sql string, args ...any) (result []Row, err error) {
	stmt, err := db.Prepare(_sql)
	if err != nil {
		return nil, errors.Wrap(err, "fly.exec.Prepare err")
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, errors.Wrap(err, "fly.exec.Query err")
	}
	defer rows.Close()

	return rows2SliceMap(rows)
}

func queryTx(tx *sql.Tx, _sql string, args ...any) (result []Row, err error) {
	stmt, err := tx.Prepare(_sql)
	if err != nil {
		return nil, errors.Wrap(err, "fly.exec.Prepare err")
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, errors.Wrap(err, "fly.exec.Query err")
	}
	defer rows.Close()

	return rows2SliceMap(rows)
}

var needConvertPlaceholder = false

// SetNeedConvertPlaceholder 设置是否需将SQL语句中的?占位符按顺序替换为$1, $2, $3等
func SetNeedConvertPlaceholder(b bool) {
	needConvertPlaceholder = b
}

// convertPlaceholders 将SQL语句中的?占位符按顺序替换为$1, $2, $3等
func convertPlaceholders(sql string) string {
	if !needConvertPlaceholder {
		return sql
	}
	// Split the SQL into parts on '?'
	parts := strings.Split(sql, "?")
	// Initialize an empty slice to hold the parts with placeholders replaced
	var newParts []string
	// Loop over the parts and append the correct placeholder
	for i, part := range parts {
		newParts = append(newParts, part)
		if i < len(parts)-1 { // Avoid appending a placeholder after the last part
			newParts = append(newParts, fmt.Sprintf("$%d", i+1))
		}
	}
	// Join the parts back together
	return strings.Join(newParts, "")
}

func destination(columnTypes []*sql.ColumnType) func() []any {
	dest := make([]func() any, 0, len(columnTypes))
	for _, v := range columnTypes {
		nullable, _ := v.Nullable()
		typeName := strings.ToUpper(v.DatabaseTypeName())

		switch typeName {
		case "VARCHAR", "CHAR", "TEXT", "NVARCHAR", "LONGTEXT", "LONGBLOB", "MEDIUMTEXT", "MEDIUMBLOB", "BLOB", "TINYTEXT":
			dest = append(dest, func() any {
				return new(sql.NullString)
			})
		case "BINARY", "VARBINARY", "TINYBLOB", "BYTEA":
			// 所有二进制类型都使用RawBytes处理，不管有没有指定长度
			dest = append(dest, func() any {
				return new(sql.RawBytes)
			})
		case "UNSIGNED INT", "UNSIGNED TINYINT", "UNSIGNED INTEGER", "UNSIGNED SMALLINT", "UNSIGNED MEDIUMINT", "UNSIGNED TINYINTEGER":
			if nullable {
				dest = append(dest, func() any {
					return new(sql.NullInt64)
				})
			} else {
				dest = append(dest, func() any {
					return new(uint)
				})
			}
		case "UNSIGNED BIGINT":
			if nullable {
				dest = append(dest, func() any {
					return new(sql.NullInt64)
				})
			} else {
				dest = append(dest, func() any {
					return new(uint64)
				})
			}
		case "INT", "INT8", "TINYINT", "INTEGER", "SMALLINT", "MEDIUMINT", "TINYINTEGER":
			if nullable {
				dest = append(dest, func() any {
					return new(sql.NullInt64)
				})
			} else {
				dest = append(dest, func() any {
					return new(int)
				})
			}
		case "BIGINT":
			if nullable {
				dest = append(dest, func() any {
					return new(sql.NullInt64)
				})
			} else {
				dest = append(dest, func() any {
					return new(int64)
				})
			}
		case "DATETIME", "DATE", "TIMESTAMP", "TIME", "TIMESTAMPTZ":
			dest = append(dest, func() any {
				return new(sql.NullTime)
			})
		case "DOUBLE", "FLOAT":
			if nullable {
				dest = append(dest, func() any {
					return new(sql.NullFloat64)
				})
			} else {
				dest = append(dest, func() any {
					return new(float64)
				})
			}
		case "DECIMAL", "NUMERIC":
			// decimal类型，无论是否可为NULL，都使用NullString接收
			// 因为decimal.Decimal不支持直接从sql.Scan接口中处理NULL
			dest = append(dest, func() any {
				return new(sql.NullString)
			})
		default:
			// 对于未知类型，默认使用可为NULL的字符串类型处理
			dest = append(dest, func() any {
				return new(sql.NullString)
			})
		}
	}
	return func() []any {
		tmp := make([]any, 0, len(dest))
		for _, d := range dest {
			tmp = append(tmp, d())
		}
		return tmp
	}
}

func rows2SliceMap(rows *sql.Rows) (list []Row, err error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, errors.Wrap(err, "fly.rows2SliceMap.columns err")
	}
	length := len(columns)

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.Wrap(err, "fly.rows2SliceMap.ColumnTypes err")
	}

	dest := destination(columnTypes)

	for rows.Next() {
		tmp := dest()
		err = rows.Scan(tmp...)
		if err != nil {
			return nil, errors.Wrap(err, "fly.rows2SliceMap.Scan err")
		}
		row := new(Row)
		row.Data = map[string]any{}
		for i := 0; i < length; i++ {
			typeName := strings.ToUpper(columnTypes[i].DatabaseTypeName())
			switch v := tmp[i].(type) {
			case *sql.NullString:
				if v.Valid {
					// 如果这是DECIMAL类型且值有效，则转换为decimal.Decimal
					if typeName == "DECIMAL" || typeName == "NUMERIC" {
						dec, err := decimal.NewFromString(v.String)
						if err == nil {
							row.Data[columns[i]] = dec
						} else {
							row.Data[columns[i]] = v.String
						}
					} else {
						row.Data[columns[i]] = v.String
					}
				} else {
					row.Data[columns[i]] = nil
				}
			case *sql.NullTime:
				if v.Valid {
					row.Data[columns[i]] = v.Time
				} else {
					row.Data[columns[i]] = nil
				}
			case *sql.NullInt64:
				if v.Valid {
					// 根据原始列类型决定返回什么类型
					if strings.HasPrefix(typeName, "UNSIGNED") {
						if strings.Contains(typeName, "BIGINT") {
							row.Data[columns[i]] = uint64(v.Int64)
						} else {
							row.Data[columns[i]] = uint(v.Int64)
						}
					} else {
						if strings.Contains(typeName, "BIGINT") {
							row.Data[columns[i]] = v.Int64
						} else {
							row.Data[columns[i]] = int(v.Int64)
						}
					}
				} else {
					row.Data[columns[i]] = nil
				}
			case *sql.NullFloat64:
				if v.Valid {
					row.Data[columns[i]] = v.Float64
				} else {
					row.Data[columns[i]] = nil
				}
			case *sql.RawBytes:
				if v != nil && len(*v) > 0 {
					// 复制字节数组，避免被后续扫描覆盖
					bytesCopy := make([]byte, len(*v))
					copy(bytesCopy, *v)
					row.Data[columns[i]] = bytesCopy
				} else {
					row.Data[columns[i]] = nil
				}
			default:
				row.Data[columns[i]] = tmp[i]
			}
		}
		list = append(list, *row)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "fly.rows2SliceMap.rows.Err err")
	}
	return list, nil
}

func SqlRows2Record(rows *sql.Rows) (list []Record, err error) {
	row, err := rows2SliceMap(rows)
	if err != nil {
		return nil, err
	}
	for _, v := range row {
		list = append(list, v.Data)
	}
	return list, nil
}

func SqlRows2SingleRecord(rows *sql.Rows) (record Record, err error) {
	row, err := rows2SliceMap(rows)
	if err != nil {
		return nil, err
	}
	for _, v := range row {
		return v.Data, nil
	}
	return nil, nil
}
