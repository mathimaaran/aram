/* Niraluli Tamil-0.62: shared SQL vtable with a dynamically loaded SQLite backend. */
#include <dlfcn.h>

#define ULI_DB_CONN_MAGIC UINT64_C(0x4152414d4442434e)
#define ULI_DB_STMT_MAGIC UINT64_C(0x4152414d44425354)
#define ULI_DB_OK 0
#define ULI_SQLITE_OK 0
#define ULI_SQLITE_NULL 5
#define ULI_SQLITE_ROW 100
#define ULI_SQLITE_DONE 101
#define ULI_SQLITE_OPEN_READWRITE 0x00000002
#define ULI_SQLITE_OPEN_CREATE 0x00000004

typedef struct uli_db_vtable uli_db_vtable;
typedef struct {
	uint64_t magic;
	const uli_db_vtable *vtable;
	void *impl;
} uli_db_conn;
typedef struct {
	uint64_t magic;
	uli_db_conn *owner;
	void *impl;
} uli_db_stmt;

typedef struct { int64_t value; void *err; } uli_db_i64_result;
typedef struct { double value; void *err; } uli_db_f64_result;
typedef struct { int value; void *err; } uli_db_bool_result;
typedef struct { const char *value; void *err; } uli_db_str_result;
typedef struct { uint8_t *data; int64_t len; void *err; } uli_db_blob_result;
typedef struct { int64_t affected; int64_t inserted; void *err; } uli_db_exec_result;
typedef struct { int64_t code; const char *message; } uli_db_error;

struct uli_db_vtable {
	const char *name;
	int (*open)(const char *, void **);
	int (*close)(void *);
	int (*exec)(void *, const char *, int64_t *, int64_t *);
	int (*prepare)(void *, const char *, void **);
	int (*bind_null)(void *, int);
	int (*bind_i64)(void *, int, int64_t);
	int (*bind_f64)(void *, int, double);
	int (*bind_text)(void *, int, const char *, int);
	int (*bind_blob)(void *, int, const void *, int);
	int (*step)(void *, int *);
	int (*column_count)(void *);
	const char *(*column_name)(void *, int);
	int (*column_is_null)(void *, int);
	int64_t (*column_i64)(void *, int);
	double (*column_f64)(void *, int);
	const unsigned char *(*column_text)(void *, int);
	const void *(*column_blob)(void *, int);
	int (*column_bytes)(void *, int);
	int (*reset)(void *);
	int (*clear_bindings)(void *);
	int (*finalize)(void *);
	const char *(*error_message)(void *);
};

static const char *uli_db_copy_text(const char *text, size_t n) {
	char *copy = (char *)uli_alloc(n + 1);
	if (n && text) memcpy(copy, text, n);
	copy[n] = '\0';
	return copy;
}

static void *uli_db_error_new(int64_t code, const char *message) {
	uli_db_error *err = (uli_db_error *)uli_alloc(sizeof(*err));
	err->code = code;
	if (!message) message = "database error";
	err->message = uli_db_copy_text(message, strlen(message));
	return err;
}

static uli_db_conn *uli_db_conn_get(int64_t handle) {
	uli_db_conn *conn = (uli_db_conn *)(intptr_t)handle;
	if (!conn || conn->magic != ULI_DB_CONN_MAGIC || !conn->vtable || !conn->impl) return NULL;
	return conn;
}

static uli_db_stmt *uli_db_stmt_get(int64_t handle) {
	uli_db_stmt *stmt = (uli_db_stmt *)(intptr_t)handle;
	if (!stmt || stmt->magic != ULI_DB_STMT_MAGIC || !stmt->owner ||
	    stmt->owner->magic != ULI_DB_CONN_MAGIC || !stmt->impl) return NULL;
	return stmt;
}

static void *uli_db_conn_error(uli_db_conn *conn, int code) {
	const char *message = NULL;
	if (conn && conn->vtable && conn->vtable->error_message && conn->impl)
		message = conn->vtable->error_message(conn->impl);
	return uli_db_error_new(code, message);
}

/* SQLite declarations are loaded with dlsym so generated C needs no sqlite3.h
   and no build-time -lsqlite3. libsqlite3.so.0 is required only at runtime. */
typedef void (*uli_sqlite_destructor)(void *);
typedef struct {
	void *library;
	int available;
	char load_error[256];
	int (*open_v2)(const char *, void **, int, const char *);
	int (*close_fn)(void *);
	const char *(*errmsg)(void *);
	int (*exec)(void *, const char *, int (*)(void *, int, char **, char **), void *, char **);
	void (*free_fn)(void *);
	int (*changes)(void *);
	int64_t (*last_insert_rowid)(void *);
	int (*prepare_v2)(void *, const char *, int, void **, const char **);
	int (*bind_null)(void *, int);
	int (*bind_int64)(void *, int, int64_t);
	int (*bind_double)(void *, int, double);
	int (*bind_text)(void *, int, const char *, int, uli_sqlite_destructor);
	int (*bind_blob)(void *, int, const void *, int, uli_sqlite_destructor);
	int (*step)(void *);
	int (*column_count)(void *);
	const char *(*column_name)(void *, int);
	int (*column_type)(void *, int);
	int64_t (*column_int64)(void *, int);
	double (*column_double)(void *, int);
	const unsigned char *(*column_text)(void *, int);
	const void *(*column_blob)(void *, int);
	int (*column_bytes)(void *, int);
	int (*reset)(void *);
	int (*clear_bindings)(void *);
	int (*finalize)(void *);
} uli_sqlite_api;

static uli_sqlite_api uli_sqlite;
static pthread_once_t uli_sqlite_once = PTHREAD_ONCE_INIT;

static void uli_sqlite_load(void) {
	memset(&uli_sqlite, 0, sizeof uli_sqlite);
	uli_sqlite.library = dlopen("libsqlite3.so.0", RTLD_NOW | RTLD_LOCAL);
	if (!uli_sqlite.library) uli_sqlite.library = dlopen("libsqlite3.so", RTLD_NOW | RTLD_LOCAL);
	if (!uli_sqlite.library) {
		const char *why = dlerror();
		snprintf(uli_sqlite.load_error, sizeof uli_sqlite.load_error,
		         "SQLite runtime unavailable: %s", why ? why : "libsqlite3 not found");
		return;
	}
#define ULI_SQLITE_LOAD(field, symbol) do { \
	*(void **)(&uli_sqlite.field) = dlsym(uli_sqlite.library, symbol); \
	if (!uli_sqlite.field) { \
		snprintf(uli_sqlite.load_error, sizeof uli_sqlite.load_error, \
		         "SQLite runtime missing symbol %s", symbol); \
		return; \
	} \
} while (0)
	ULI_SQLITE_LOAD(open_v2, "sqlite3_open_v2");
	ULI_SQLITE_LOAD(close_fn, "sqlite3_close");
	ULI_SQLITE_LOAD(errmsg, "sqlite3_errmsg");
	ULI_SQLITE_LOAD(exec, "sqlite3_exec");
	ULI_SQLITE_LOAD(free_fn, "sqlite3_free");
	ULI_SQLITE_LOAD(changes, "sqlite3_changes");
	ULI_SQLITE_LOAD(last_insert_rowid, "sqlite3_last_insert_rowid");
	ULI_SQLITE_LOAD(prepare_v2, "sqlite3_prepare_v2");
	ULI_SQLITE_LOAD(bind_null, "sqlite3_bind_null");
	ULI_SQLITE_LOAD(bind_int64, "sqlite3_bind_int64");
	ULI_SQLITE_LOAD(bind_double, "sqlite3_bind_double");
	ULI_SQLITE_LOAD(bind_text, "sqlite3_bind_text");
	ULI_SQLITE_LOAD(bind_blob, "sqlite3_bind_blob");
	ULI_SQLITE_LOAD(step, "sqlite3_step");
	ULI_SQLITE_LOAD(column_count, "sqlite3_column_count");
	ULI_SQLITE_LOAD(column_name, "sqlite3_column_name");
	ULI_SQLITE_LOAD(column_type, "sqlite3_column_type");
	ULI_SQLITE_LOAD(column_int64, "sqlite3_column_int64");
	ULI_SQLITE_LOAD(column_double, "sqlite3_column_double");
	ULI_SQLITE_LOAD(column_text, "sqlite3_column_text");
	ULI_SQLITE_LOAD(column_blob, "sqlite3_column_blob");
	ULI_SQLITE_LOAD(column_bytes, "sqlite3_column_bytes");
	ULI_SQLITE_LOAD(reset, "sqlite3_reset");
	ULI_SQLITE_LOAD(clear_bindings, "sqlite3_clear_bindings");
	ULI_SQLITE_LOAD(finalize, "sqlite3_finalize");
#undef ULI_SQLITE_LOAD
	uli_sqlite.available = 1;
}

static int uli_sqlite_open(const char *dsn, void **out) {
	pthread_once(&uli_sqlite_once, uli_sqlite_load);
	if (!uli_sqlite.available) return -1;
	return uli_sqlite.open_v2(dsn, out,
		ULI_SQLITE_OPEN_READWRITE | ULI_SQLITE_OPEN_CREATE, NULL);
}

static int uli_sqlite_close(void *conn) { return uli_sqlite.close_fn(conn); }

static int uli_sqlite_exec(void *conn, const char *sql, int64_t *affected, int64_t *inserted) {
	char *message = NULL;
	int rc = uli_sqlite.exec(conn, sql, NULL, NULL, &message);
	if (message) uli_sqlite.free_fn(message);
	if (rc == ULI_SQLITE_OK) {
		*affected = uli_sqlite.changes(conn);
		*inserted = uli_sqlite.last_insert_rowid(conn);
	}
	return rc;
}

static int uli_sqlite_prepare(void *conn, const char *sql, void **out) {
	return uli_sqlite.prepare_v2(conn, sql, -1, out, NULL);
}
static int uli_sqlite_bind_null(void *stmt, int index) {
	return uli_sqlite.bind_null(stmt, index);
}
static int uli_sqlite_bind_i64(void *stmt, int index, int64_t value) {
	return uli_sqlite.bind_int64(stmt, index, value);
}
static int uli_sqlite_bind_f64(void *stmt, int index, double value) {
	return uli_sqlite.bind_double(stmt, index, value);
}
static int uli_sqlite_bind_text(void *stmt, int index, const char *value, int len) {
	return uli_sqlite.bind_text(stmt, index, value, len, (uli_sqlite_destructor)(intptr_t)-1);
}
static int uli_sqlite_bind_blob(void *stmt, int index, const void *value, int len) {
	return uli_sqlite.bind_blob(stmt, index, value, len, (uli_sqlite_destructor)(intptr_t)-1);
}
static int uli_sqlite_step(void *stmt, int *has_row) {
	int rc = uli_sqlite.step(stmt);
	if (rc == ULI_SQLITE_ROW) {
		*has_row = 1;
		return ULI_DB_OK;
	}
	if (rc == ULI_SQLITE_DONE) {
		*has_row = 0;
		return ULI_DB_OK;
	}
	return rc;
}
static int uli_sqlite_column_count(void *stmt) { return uli_sqlite.column_count(stmt); }
static const char *uli_sqlite_column_name(void *stmt, int index) {
	return uli_sqlite.column_name(stmt, index);
}
static int uli_sqlite_column_is_null(void *stmt, int index) {
	return uli_sqlite.column_type(stmt, index) == ULI_SQLITE_NULL;
}
static int64_t uli_sqlite_column_i64(void *stmt, int index) {
	return uli_sqlite.column_int64(stmt, index);
}
static double uli_sqlite_column_f64(void *stmt, int index) {
	return uli_sqlite.column_double(stmt, index);
}
static const unsigned char *uli_sqlite_column_text(void *stmt, int index) {
	return uli_sqlite.column_text(stmt, index);
}
static const void *uli_sqlite_column_blob(void *stmt, int index) {
	return uli_sqlite.column_blob(stmt, index);
}
static int uli_sqlite_column_bytes(void *stmt, int index) {
	return uli_sqlite.column_bytes(stmt, index);
}
static int uli_sqlite_reset(void *stmt) { return uli_sqlite.reset(stmt); }
static int uli_sqlite_clear_bindings(void *stmt) { return uli_sqlite.clear_bindings(stmt); }
static int uli_sqlite_finalize(void *stmt) { return uli_sqlite.finalize(stmt); }
static const char *uli_sqlite_error_message(void *conn) { return uli_sqlite.errmsg(conn); }

static const uli_db_vtable uli_sqlite_vtable = {
	"sqlite",
	uli_sqlite_open, uli_sqlite_close, uli_sqlite_exec, uli_sqlite_prepare,
	uli_sqlite_bind_null, uli_sqlite_bind_i64, uli_sqlite_bind_f64,
	uli_sqlite_bind_text, uli_sqlite_bind_blob, uli_sqlite_step,
	uli_sqlite_column_count, uli_sqlite_column_name, uli_sqlite_column_is_null,
	uli_sqlite_column_i64, uli_sqlite_column_f64, uli_sqlite_column_text,
	uli_sqlite_column_blob, uli_sqlite_column_bytes, uli_sqlite_reset,
	uli_sqlite_clear_bindings, uli_sqlite_finalize, uli_sqlite_error_message
};

static const uli_db_vtable *uli_db_backend(const char *driver) {
	if (driver && (strcmp(driver, "sqlite") == 0 || strcmp(driver, "sqlite3") == 0))
		return &uli_sqlite_vtable;
	return NULL;
}

static uli_db_i64_result uli_db_open(const char *driver, const char *dsn) {
	const uli_db_vtable *vtable = uli_db_backend(driver);
	if (!vtable)
		return (uli_db_i64_result){ 0, uli_db_error_new(-2, "unsupported database driver") };
	if (!dsn || !*dsn)
		return (uli_db_i64_result){ 0, uli_db_error_new(-3, "empty database address") };
	void *impl = NULL;
	char sp;
	uli_gc_park_begin(&sp);
	int rc = vtable->open(dsn, &impl);
	uli_gc_park_end();
	if (rc != ULI_DB_OK) {
		const char *message = NULL;
		if (impl && vtable->error_message) message = vtable->error_message(impl);
		if (!message && uli_sqlite.load_error[0]) message = uli_sqlite.load_error;
		void *err = uli_db_error_new(rc, message);
		if (impl) vtable->close(impl);
		return (uli_db_i64_result){ 0, err };
	}
	uli_db_conn *conn = (uli_db_conn *)uli_alloc(sizeof(*conn));
	conn->magic = ULI_DB_CONN_MAGIC;
	conn->vtable = vtable;
	conn->impl = impl;
	return (uli_db_i64_result){ (int64_t)(intptr_t)conn, NULL };
}

static void *uli_db_close(int64_t handle) {
	volatile uli_db_conn *root = uli_db_conn_get(handle);
	if (!root) return uli_db_error_new(-4, "invalid or closed database handle");
	char sp;
	uli_gc_park_begin(&sp);
	int rc = root->vtable->close(root->impl);
	uli_gc_park_end();
	if (rc != ULI_DB_OK) return uli_db_conn_error((uli_db_conn *)root, rc);
	root->impl = NULL;
	root->magic = 0;
	return NULL;
}

static uli_db_exec_result uli_db_exec(int64_t handle, const char *sql) {
	volatile uli_db_conn *root = uli_db_conn_get(handle);
	if (!root) return (uli_db_exec_result){ 0, 0, uli_db_error_new(-4, "invalid or closed database handle") };
	if (!sql) return (uli_db_exec_result){ 0, 0, uli_db_error_new(-5, "nil SQL text") };
	int64_t affected = 0, inserted = 0;
	char sp;
	uli_gc_park_begin(&sp);
	int rc = root->vtable->exec(root->impl, sql, &affected, &inserted);
	uli_gc_park_end();
	if (rc != ULI_DB_OK)
		return (uli_db_exec_result){ 0, 0, uli_db_conn_error((uli_db_conn *)root, rc) };
	return (uli_db_exec_result){ affected, inserted, NULL };
}

static uli_db_i64_result uli_db_prepare(int64_t handle, const char *sql) {
	volatile uli_db_conn *root = uli_db_conn_get(handle);
	if (!root) return (uli_db_i64_result){ 0, uli_db_error_new(-4, "invalid or closed database handle") };
	void *impl = NULL;
	char sp;
	uli_gc_park_begin(&sp);
	int rc = root->vtable->prepare(root->impl, sql, &impl);
	uli_gc_park_end();
	if (rc != ULI_DB_OK)
		return (uli_db_i64_result){ 0, uli_db_conn_error((uli_db_conn *)root, rc) };
	uli_db_stmt *stmt = (uli_db_stmt *)uli_alloc(sizeof(*stmt));
	stmt->magic = ULI_DB_STMT_MAGIC;
	stmt->owner = (uli_db_conn *)root;
	stmt->impl = impl;
	return (uli_db_i64_result){ (int64_t)(intptr_t)stmt, NULL };
}

static void *uli_db_bind_null(int64_t handle, int64_t index) {
	volatile uli_db_stmt *root = uli_db_stmt_get(handle);
	if (!root) return uli_db_error_new(-6, "invalid or finalized statement handle");
	int rc = root->owner->vtable->bind_null(root->impl, (int)index);
	return rc == ULI_DB_OK ? NULL : uli_db_conn_error(root->owner, rc);
}
static void *uli_db_bind_i64(int64_t handle, int64_t index, int64_t value) {
	volatile uli_db_stmt *root = uli_db_stmt_get(handle);
	if (!root) return uli_db_error_new(-6, "invalid or finalized statement handle");
	int rc = root->owner->vtable->bind_i64(root->impl, (int)index, value);
	return rc == ULI_DB_OK ? NULL : uli_db_conn_error(root->owner, rc);
}
static void *uli_db_bind_f64(int64_t handle, int64_t index, double value) {
	volatile uli_db_stmt *root = uli_db_stmt_get(handle);
	if (!root) return uli_db_error_new(-6, "invalid or finalized statement handle");
	int rc = root->owner->vtable->bind_f64(root->impl, (int)index, value);
	return rc == ULI_DB_OK ? NULL : uli_db_conn_error(root->owner, rc);
}
static void *uli_db_bind_text(int64_t handle, int64_t index, const char *value) {
	volatile uli_db_stmt *root = uli_db_stmt_get(handle);
	if (!root) return uli_db_error_new(-6, "invalid or finalized statement handle");
	size_t n = value ? strlen(value) : 0;
	if (n > INT32_MAX) return uli_db_error_new(-7, "text parameter too large");
	int rc = root->owner->vtable->bind_text(root->impl, (int)index, value ? value : "", (int)n);
	return rc == ULI_DB_OK ? NULL : uli_db_conn_error(root->owner, rc);
}
static void *uli_db_bind_blob(int64_t handle, int64_t index, const uint8_t *data, int64_t len) {
	volatile uli_db_stmt *root = uli_db_stmt_get(handle);
	if (!root) return uli_db_error_new(-6, "invalid or finalized statement handle");
	if (len < 0 || len > INT32_MAX) return uli_db_error_new(-7, "blob parameter too large");
	static const uint8_t empty_blob = 0;
	const uint8_t *value = data ? data : &empty_blob;
	int rc = root->owner->vtable->bind_blob(root->impl, (int)index, value, (int)len);
	return rc == ULI_DB_OK ? NULL : uli_db_conn_error(root->owner, rc);
}

static uli_db_bool_result uli_db_step(int64_t handle) {
	volatile uli_db_stmt *root = uli_db_stmt_get(handle);
	if (!root) return (uli_db_bool_result){ 0, uli_db_error_new(-6, "invalid or finalized statement handle") };
	char sp;
	uli_gc_park_begin(&sp);
	int has_row = 0;
	int rc = root->owner->vtable->step(root->impl, &has_row);
	uli_gc_park_end();
	if (rc == ULI_DB_OK) return (uli_db_bool_result){ has_row, NULL };
	return (uli_db_bool_result){ 0, uli_db_conn_error(root->owner, rc) };
}

static int uli_db_column_valid(uli_db_stmt *stmt, int64_t index) {
	return stmt && index >= 0 && index < stmt->owner->vtable->column_count(stmt->impl);
}

static uli_db_i64_result uli_db_column_count(int64_t handle) {
	uli_db_stmt *stmt = uli_db_stmt_get(handle);
	if (!stmt)
		return (uli_db_i64_result){ 0, uli_db_error_new(-6, "invalid or finalized statement handle") };
	return (uli_db_i64_result){ stmt->owner->vtable->column_count(stmt->impl), NULL };
}

static uli_db_str_result uli_db_column_name(int64_t handle, int64_t index) {
	uli_db_stmt *stmt = uli_db_stmt_get(handle);
	if (!stmt) return (uli_db_str_result){ "", uli_db_error_new(-6, "invalid or finalized statement handle") };
	if (!uli_db_column_valid(stmt, index))
		return (uli_db_str_result){ "", uli_db_error_new(-8, "column index out of range") };
	const char *name = stmt->owner->vtable->column_name(stmt->impl, (int)index);
	return (uli_db_str_result){ uli_db_copy_text(name ? name : "", name ? strlen(name) : 0), NULL };
}

static uli_db_bool_result uli_db_column_is_null(int64_t handle, int64_t index) {
	uli_db_stmt *stmt = uli_db_stmt_get(handle);
	if (!stmt) return (uli_db_bool_result){ 0, uli_db_error_new(-6, "invalid or finalized statement handle") };
	if (!uli_db_column_valid(stmt, index))
		return (uli_db_bool_result){ 0, uli_db_error_new(-8, "column index out of range") };
	return (uli_db_bool_result){
		stmt->owner->vtable->column_is_null(stmt->impl, (int)index), NULL
	};
}

static uli_db_i64_result uli_db_column_i64(int64_t handle, int64_t index) {
	uli_db_stmt *stmt = uli_db_stmt_get(handle);
	if (!stmt) return (uli_db_i64_result){ 0, uli_db_error_new(-6, "invalid or finalized statement handle") };
	if (!uli_db_column_valid(stmt, index))
		return (uli_db_i64_result){ 0, uli_db_error_new(-8, "column index out of range") };
	return (uli_db_i64_result){ stmt->owner->vtable->column_i64(stmt->impl, (int)index), NULL };
}

static uli_db_f64_result uli_db_column_f64(int64_t handle, int64_t index) {
	uli_db_stmt *stmt = uli_db_stmt_get(handle);
	if (!stmt) return (uli_db_f64_result){ 0, uli_db_error_new(-6, "invalid or finalized statement handle") };
	if (!uli_db_column_valid(stmt, index))
		return (uli_db_f64_result){ 0, uli_db_error_new(-8, "column index out of range") };
	return (uli_db_f64_result){ stmt->owner->vtable->column_f64(stmt->impl, (int)index), NULL };
}

static uli_db_str_result uli_db_column_text(int64_t handle, int64_t index) {
	uli_db_stmt *stmt = uli_db_stmt_get(handle);
	if (!stmt) return (uli_db_str_result){ "", uli_db_error_new(-6, "invalid or finalized statement handle") };
	if (!uli_db_column_valid(stmt, index))
		return (uli_db_str_result){ "", uli_db_error_new(-8, "column index out of range") };
	const unsigned char *text = stmt->owner->vtable->column_text(stmt->impl, (int)index);
	int n = stmt->owner->vtable->column_bytes(stmt->impl, (int)index);
	return (uli_db_str_result){ uli_db_copy_text((const char *)text, text ? (size_t)n : 0), NULL };
}

static uli_db_blob_result uli_db_column_blob(int64_t handle, int64_t index) {
	uli_db_stmt *stmt = uli_db_stmt_get(handle);
	if (!stmt) return (uli_db_blob_result){ NULL, 0, uli_db_error_new(-6, "invalid or finalized statement handle") };
	if (!uli_db_column_valid(stmt, index))
		return (uli_db_blob_result){ NULL, 0, uli_db_error_new(-8, "column index out of range") };
	const void *data = stmt->owner->vtable->column_blob(stmt->impl, (int)index);
	int n = stmt->owner->vtable->column_bytes(stmt->impl, (int)index);
	if (!data || n <= 0) return (uli_db_blob_result){ NULL, 0, NULL };
	uint8_t *copy = (uint8_t *)uli_alloc((size_t)n);
	memcpy(copy, data, (size_t)n);
	return (uli_db_blob_result){ copy, n, NULL };
}

static void *uli_db_reset(int64_t handle) {
	uli_db_stmt *stmt = uli_db_stmt_get(handle);
	if (!stmt) return uli_db_error_new(-6, "invalid or finalized statement handle");
	int rc = stmt->owner->vtable->reset(stmt->impl);
	if (rc != ULI_DB_OK) return uli_db_conn_error(stmt->owner, rc);
	rc = stmt->owner->vtable->clear_bindings(stmt->impl);
	return rc == ULI_DB_OK ? NULL : uli_db_conn_error(stmt->owner, rc);
}

static void *uli_db_finalize(int64_t handle) {
	volatile uli_db_stmt *root = uli_db_stmt_get(handle);
	if (!root) return uli_db_error_new(-6, "invalid or finalized statement handle");
	uli_db_conn *owner = root->owner;
	char sp;
	uli_gc_park_begin(&sp);
	int rc = owner->vtable->finalize(root->impl);
	uli_gc_park_end();
	root->impl = NULL;
	root->magic = 0;
	return rc == ULI_DB_OK ? NULL : uli_db_conn_error(owner, rc);
}
