// Code generated by SQLBoiler (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/queries/qmhelper"
	"github.com/volatiletech/sqlboiler/strmangle"
)

// ReputationLog is an object representing the database table.
type ReputationLog struct {
	ID               int64     `boil:"id" json:"id" toml:"id" yaml:"id"`
	CreatedAt        time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	GuildID          int64     `boil:"guild_id" json:"guild_id" toml:"guild_id" yaml:"guild_id"`
	SenderID         int64     `boil:"sender_id" json:"sender_id" toml:"sender_id" yaml:"sender_id"`
	ReceiverID       int64     `boil:"receiver_id" json:"receiver_id" toml:"receiver_id" yaml:"receiver_id"`
	SetFixedAmount   bool      `boil:"set_fixed_amount" json:"set_fixed_amount" toml:"set_fixed_amount" yaml:"set_fixed_amount"`
	Amount           int64     `boil:"amount" json:"amount" toml:"amount" yaml:"amount"`
	ReceiverUsername string    `boil:"receiver_username" json:"receiver_username" toml:"receiver_username" yaml:"receiver_username"`
	SenderUsername   string    `boil:"sender_username" json:"sender_username" toml:"sender_username" yaml:"sender_username"`

	R *reputationLogR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L reputationLogL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var ReputationLogColumns = struct {
	ID               string
	CreatedAt        string
	GuildID          string
	SenderID         string
	ReceiverID       string
	SetFixedAmount   string
	Amount           string
	ReceiverUsername string
	SenderUsername   string
}{
	ID:               "id",
	CreatedAt:        "created_at",
	GuildID:          "guild_id",
	SenderID:         "sender_id",
	ReceiverID:       "receiver_id",
	SetFixedAmount:   "set_fixed_amount",
	Amount:           "amount",
	ReceiverUsername: "receiver_username",
	SenderUsername:   "sender_username",
}

// Generated where

type whereHelpertime_Time struct{ field string }

func (w whereHelpertime_Time) EQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertime_Time) NEQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertime_Time) LT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertime_Time) LTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertime_Time) GT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertime_Time) GTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var ReputationLogWhere = struct {
	ID               whereHelperint64
	CreatedAt        whereHelpertime_Time
	GuildID          whereHelperint64
	SenderID         whereHelperint64
	ReceiverID       whereHelperint64
	SetFixedAmount   whereHelperbool
	Amount           whereHelperint64
	ReceiverUsername whereHelperstring
	SenderUsername   whereHelperstring
}{
	ID:               whereHelperint64{field: "\"reputation_log\".\"id\""},
	CreatedAt:        whereHelpertime_Time{field: "\"reputation_log\".\"created_at\""},
	GuildID:          whereHelperint64{field: "\"reputation_log\".\"guild_id\""},
	SenderID:         whereHelperint64{field: "\"reputation_log\".\"sender_id\""},
	ReceiverID:       whereHelperint64{field: "\"reputation_log\".\"receiver_id\""},
	SetFixedAmount:   whereHelperbool{field: "\"reputation_log\".\"set_fixed_amount\""},
	Amount:           whereHelperint64{field: "\"reputation_log\".\"amount\""},
	ReceiverUsername: whereHelperstring{field: "\"reputation_log\".\"receiver_username\""},
	SenderUsername:   whereHelperstring{field: "\"reputation_log\".\"sender_username\""},
}

// ReputationLogRels is where relationship names are stored.
var ReputationLogRels = struct {
}{}

// reputationLogR is where relationships are stored.
type reputationLogR struct {
}

// NewStruct creates a new relationship struct
func (*reputationLogR) NewStruct() *reputationLogR {
	return &reputationLogR{}
}

// reputationLogL is where Load methods for each relationship are stored.
type reputationLogL struct{}

var (
	reputationLogAllColumns            = []string{"id", "created_at", "guild_id", "sender_id", "receiver_id", "set_fixed_amount", "amount", "receiver_username", "sender_username"}
	reputationLogColumnsWithoutDefault = []string{"created_at", "guild_id", "sender_id", "receiver_id", "set_fixed_amount", "amount"}
	reputationLogColumnsWithDefault    = []string{"id", "receiver_username", "sender_username"}
	reputationLogPrimaryKeyColumns     = []string{"id"}
)

type (
	// ReputationLogSlice is an alias for a slice of pointers to ReputationLog.
	// This should generally be used opposed to []ReputationLog.
	ReputationLogSlice []*ReputationLog

	reputationLogQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	reputationLogType                 = reflect.TypeOf(&ReputationLog{})
	reputationLogMapping              = queries.MakeStructMapping(reputationLogType)
	reputationLogPrimaryKeyMapping, _ = queries.BindMapping(reputationLogType, reputationLogMapping, reputationLogPrimaryKeyColumns)
	reputationLogInsertCacheMut       sync.RWMutex
	reputationLogInsertCache          = make(map[string]insertCache)
	reputationLogUpdateCacheMut       sync.RWMutex
	reputationLogUpdateCache          = make(map[string]updateCache)
	reputationLogUpsertCacheMut       sync.RWMutex
	reputationLogUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// OneG returns a single reputationLog record from the query using the global executor.
func (q reputationLogQuery) OneG(ctx context.Context) (*ReputationLog, error) {
	return q.One(ctx, boil.GetContextDB())
}

// One returns a single reputationLog record from the query.
func (q reputationLogQuery) One(ctx context.Context, exec boil.ContextExecutor) (*ReputationLog, error) {
	o := &ReputationLog{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.WrapIf(err, "models: failed to execute a one query for reputation_log")
	}

	return o, nil
}

// AllG returns all ReputationLog records from the query using the global executor.
func (q reputationLogQuery) AllG(ctx context.Context) (ReputationLogSlice, error) {
	return q.All(ctx, boil.GetContextDB())
}

// All returns all ReputationLog records from the query.
func (q reputationLogQuery) All(ctx context.Context, exec boil.ContextExecutor) (ReputationLogSlice, error) {
	var o []*ReputationLog

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.WrapIf(err, "models: failed to assign all query results to ReputationLog slice")
	}

	return o, nil
}

// CountG returns the count of all ReputationLog records in the query, and panics on error.
func (q reputationLogQuery) CountG(ctx context.Context) (int64, error) {
	return q.Count(ctx, boil.GetContextDB())
}

// Count returns the count of all ReputationLog records in the query.
func (q reputationLogQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.WrapIf(err, "models: failed to count reputation_log rows")
	}

	return count, nil
}

// ExistsG checks if the row exists in the table, and panics on error.
func (q reputationLogQuery) ExistsG(ctx context.Context) (bool, error) {
	return q.Exists(ctx, boil.GetContextDB())
}

// Exists checks if the row exists in the table.
func (q reputationLogQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.WrapIf(err, "models: failed to check if reputation_log exists")
	}

	return count > 0, nil
}

// ReputationLogs retrieves all the records using an executor.
func ReputationLogs(mods ...qm.QueryMod) reputationLogQuery {
	mods = append(mods, qm.From("\"reputation_log\""))
	return reputationLogQuery{NewQuery(mods...)}
}

// FindReputationLogG retrieves a single record by ID.
func FindReputationLogG(ctx context.Context, iD int64, selectCols ...string) (*ReputationLog, error) {
	return FindReputationLog(ctx, boil.GetContextDB(), iD, selectCols...)
}

// FindReputationLog retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindReputationLog(ctx context.Context, exec boil.ContextExecutor, iD int64, selectCols ...string) (*ReputationLog, error) {
	reputationLogObj := &ReputationLog{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"reputation_log\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, reputationLogObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.WrapIf(err, "models: unable to select from reputation_log")
	}

	return reputationLogObj, nil
}

// InsertG a single record. See Insert for whitelist behavior description.
func (o *ReputationLog) InsertG(ctx context.Context, columns boil.Columns) error {
	return o.Insert(ctx, boil.GetContextDB(), columns)
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *ReputationLog) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no reputation_log provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(reputationLogColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	reputationLogInsertCacheMut.RLock()
	cache, cached := reputationLogInsertCache[key]
	reputationLogInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			reputationLogAllColumns,
			reputationLogColumnsWithDefault,
			reputationLogColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(reputationLogType, reputationLogMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(reputationLogType, reputationLogMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"reputation_log\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"reputation_log\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.WrapIf(err, "models: unable to insert into reputation_log")
	}

	if !cached {
		reputationLogInsertCacheMut.Lock()
		reputationLogInsertCache[key] = cache
		reputationLogInsertCacheMut.Unlock()
	}

	return nil
}

// UpdateG a single ReputationLog record using the global executor.
// See Update for more documentation.
func (o *ReputationLog) UpdateG(ctx context.Context, columns boil.Columns) (int64, error) {
	return o.Update(ctx, boil.GetContextDB(), columns)
}

// Update uses an executor to update the ReputationLog.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *ReputationLog) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	key := makeCacheKey(columns, nil)
	reputationLogUpdateCacheMut.RLock()
	cache, cached := reputationLogUpdateCache[key]
	reputationLogUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			reputationLogAllColumns,
			reputationLogPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update reputation_log, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"reputation_log\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, reputationLogPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(reputationLogType, reputationLogMapping, append(wl, reputationLogPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, values)
	}

	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.WrapIf(err, "models: unable to update reputation_log row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.WrapIf(err, "models: failed to get rows affected by update for reputation_log")
	}

	if !cached {
		reputationLogUpdateCacheMut.Lock()
		reputationLogUpdateCache[key] = cache
		reputationLogUpdateCacheMut.Unlock()
	}

	return rowsAff, nil
}

// UpdateAllG updates all rows with the specified column values.
func (q reputationLogQuery) UpdateAllG(ctx context.Context, cols M) (int64, error) {
	return q.UpdateAll(ctx, boil.GetContextDB(), cols)
}

// UpdateAll updates all rows with the specified column values.
func (q reputationLogQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.WrapIf(err, "models: unable to update all for reputation_log")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.WrapIf(err, "models: unable to retrieve rows affected for reputation_log")
	}

	return rowsAff, nil
}

// UpdateAllG updates all rows with the specified column values.
func (o ReputationLogSlice) UpdateAllG(ctx context.Context, cols M) (int64, error) {
	return o.UpdateAll(ctx, boil.GetContextDB(), cols)
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o ReputationLogSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), reputationLogPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"reputation_log\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, reputationLogPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.WrapIf(err, "models: unable to update all in reputationLog slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.WrapIf(err, "models: unable to retrieve rows affected all in update all reputationLog")
	}
	return rowsAff, nil
}

// UpsertG attempts an insert, and does an update or ignore on conflict.
func (o *ReputationLog) UpsertG(ctx context.Context, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	return o.Upsert(ctx, boil.GetContextDB(), updateOnConflict, conflictColumns, updateColumns, insertColumns)
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *ReputationLog) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no reputation_log provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(reputationLogColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	reputationLogUpsertCacheMut.RLock()
	cache, cached := reputationLogUpsertCache[key]
	reputationLogUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			reputationLogAllColumns,
			reputationLogColumnsWithDefault,
			reputationLogColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			reputationLogAllColumns,
			reputationLogPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert reputation_log, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(reputationLogPrimaryKeyColumns))
			copy(conflict, reputationLogPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"reputation_log\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(reputationLogType, reputationLogMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(reputationLogType, reputationLogMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if err == sql.ErrNoRows {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.WrapIf(err, "models: unable to upsert reputation_log")
	}

	if !cached {
		reputationLogUpsertCacheMut.Lock()
		reputationLogUpsertCache[key] = cache
		reputationLogUpsertCacheMut.Unlock()
	}

	return nil
}

// DeleteG deletes a single ReputationLog record.
// DeleteG will match against the primary key column to find the record to delete.
func (o *ReputationLog) DeleteG(ctx context.Context) (int64, error) {
	return o.Delete(ctx, boil.GetContextDB())
}

// Delete deletes a single ReputationLog record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *ReputationLog) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no ReputationLog provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), reputationLogPrimaryKeyMapping)
	sql := "DELETE FROM \"reputation_log\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.WrapIf(err, "models: unable to delete from reputation_log")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.WrapIf(err, "models: failed to get rows affected by delete for reputation_log")
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q reputationLogQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no reputationLogQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.WrapIf(err, "models: unable to delete all from reputation_log")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.WrapIf(err, "models: failed to get rows affected by deleteall for reputation_log")
	}

	return rowsAff, nil
}

// DeleteAllG deletes all rows in the slice.
func (o ReputationLogSlice) DeleteAllG(ctx context.Context) (int64, error) {
	return o.DeleteAll(ctx, boil.GetContextDB())
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o ReputationLogSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), reputationLogPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"reputation_log\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, reputationLogPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}

	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.WrapIf(err, "models: unable to delete all from reputationLog slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.WrapIf(err, "models: failed to get rows affected by deleteall for reputation_log")
	}

	return rowsAff, nil
}

// ReloadG refetches the object from the database using the primary keys.
func (o *ReputationLog) ReloadG(ctx context.Context) error {
	if o == nil {
		return errors.New("models: no ReputationLog provided for reload")
	}

	return o.Reload(ctx, boil.GetContextDB())
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *ReputationLog) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindReputationLog(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAllG refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ReputationLogSlice) ReloadAllG(ctx context.Context) error {
	if o == nil {
		return errors.New("models: empty ReputationLogSlice provided for reload all")
	}

	return o.ReloadAll(ctx, boil.GetContextDB())
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ReputationLogSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := ReputationLogSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), reputationLogPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"reputation_log\".* FROM \"reputation_log\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, reputationLogPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.WrapIf(err, "models: unable to reload all in ReputationLogSlice")
	}

	*o = slice

	return nil
}

// ReputationLogExistsG checks if the ReputationLog row exists.
func ReputationLogExistsG(ctx context.Context, iD int64) (bool, error) {
	return ReputationLogExists(ctx, boil.GetContextDB(), iD)
}

// ReputationLogExists checks if the ReputationLog row exists.
func ReputationLogExists(ctx context.Context, exec boil.ContextExecutor, iD int64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"reputation_log\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}

	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.WrapIf(err, "models: unable to check if reputation_log exists")
	}

	return exists, nil
}
