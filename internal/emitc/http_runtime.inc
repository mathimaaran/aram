/* Aram Tamil-0.56: minimal HTTP/1.0 server (Linux). Uses aram_net_* + GC park. */

typedef struct {
	const char *method;
	const char *path;
	aram_slice_u8 body;
	void *headers;
} aram_http_req;

typedef void (*aram_http_handler_fn)(void *env, aram_http_req req, int64_t w);

typedef struct {
	const char *path;
	aram_fn handler;
} aram_http_route;

typedef struct {
	aram_http_route *routes;
	int64_t len;
	int64_t cap;
} aram_http_mux;

static int64_t aram_http_mux_new(void) {
	aram_http_mux *mux = (aram_http_mux *)aram_alloc(sizeof(*mux));
	mux->cap = 8;
	mux->routes = (aram_http_route *)aram_alloc((size_t)mux->cap * sizeof(*mux->routes));
	return (int64_t)(intptr_t)mux;
}

static void aram_http_mux_handle(int64_t handle, const char *path, aram_fn handler) {
	aram_http_mux *mux = (aram_http_mux *)(intptr_t)handle;
	if (!mux || !path || !handler.fn) abort();
	for (int64_t i = 0; i < mux->len; i++) {
		if (strcmp(mux->routes[i].path, path) == 0) {
			mux->routes[i].handler = handler;
			return;
		}
	}
	if (mux->len == mux->cap) {
		int64_t cap = mux->cap * 2;
		aram_http_route *routes = (aram_http_route *)aram_alloc((size_t)cap * sizeof(*routes));
		memcpy(routes, mux->routes, (size_t)mux->len * sizeof(*routes));
		mux->routes = routes;
		mux->cap = cap;
	}
	mux->routes[mux->len++] = (aram_http_route){ path, handler };
}

static const char *aram_http_status_text(int code) {
	switch (code) {
	case 200: return "OK";
	case 201: return "Created";
	case 204: return "No Content";
	case 400: return "Bad Request";
	case 404: return "Not Found";
	case 405: return "Method Not Allowed";
	case 500: return "Internal Server Error";
	default: return "";
	}
}

static int64_t aram_http_header_i64(const char *headers, const char *key) {
	size_t klen = strlen(key);
	const char *p = headers;
	while (*p) {
		const char *eol = strstr(p, "\r\n");
		if (!eol) break;
		if ((size_t)(eol - p) > klen + 1 &&
		    strncasecmp(p, key, klen) == 0 && p[klen] == ':') {
			const char *v = p + klen + 1;
			while (*v == ' ' || *v == '\t') v++;
			return (int64_t)strtoll(v, NULL, 10);
		}
		if (eol[0] == '\0') break;
		p = eol + 2;
	}
	return -1;
}

static int aram_http_read_req(int64_t fd, aram_http_req *out, void **err) {
	size_t cap = 4096;
	char *buf = (char *)aram_alloc(cap);
	size_t n = 0;
	char *hdr_end = NULL;
	for (;;) {
		if (n + 512 >= cap) {
			if (cap >= 65536) {
				*err = aram_net_error_new(EOVERFLOW, "HTTP headers too large");
				return 0;
			}
			size_t ncap = cap * 2;
			char *nb = (char *)aram_alloc(ncap);
			memcpy(nb, buf, n);
			buf = nb;
			cap = ncap;
		}
		char sp;
		aram_gc_park_begin(&sp);
		ssize_t r = read((int)fd, buf + n, cap - n - 1);
		aram_gc_park_end();
		if (r < 0) {
			*err = aram_net_errno_error();
			return 0;
		}
		if (r == 0) break;
		n += (size_t)r;
		buf[n] = '\0';
		hdr_end = strstr(buf, "\r\n\r\n");
		if (!hdr_end) hdr_end = strstr(buf, "\n\n");
		if (hdr_end) break;
	}
	if (!hdr_end) {
		*err = aram_net_error_new(EINVAL, "malformed HTTP request");
		return 0;
	}
	size_t hdr_len = (size_t)(hdr_end - buf) + 4;
	char *line_end = strstr(buf, "\r\n");
	if (!line_end) {
		*err = aram_net_error_new(EINVAL, "malformed HTTP request");
		return 0;
	}
	*line_end = '\0';
	char *method = buf;
	char *path = strchr(buf, ' ');
	if (!path) {
		*err = aram_net_error_new(EINVAL, "malformed HTTP request");
		return 0;
	}
	*path++ = '\0';
	char *version = strchr(path, ' ');
	if (!version) {
		*err = aram_net_error_new(EINVAL, "malformed HTTP request");
		return 0;
	}
	*version++ = '\0';
	(void)version;
	int64_t cl = aram_http_header_i64(line_end + 2, "Content-Length");
	void *headers = aram_http_headers_new();
	char *header = line_end + 2;
	while (header < hdr_end) {
		char *eol = strstr(header, "\r\n");
		if (!eol || eol > hdr_end) break;
		char *colon = memchr(header, ':', (size_t)(eol - header));
		if (colon) {
			*colon = '\0';
			*eol = '\0';
			for (char *p = header; *p; p++) {
				*p = (char)tolower((unsigned char)*p);
			}
			char *value = colon + 1;
			while (*value == ' ' || *value == '\t') value++;
			char *end = eol;
			while (end > value && (end[-1] == ' ' || end[-1] == '\t')) *--end = '\0';
			aram_http_headers_set(headers, header, value);
		}
		header = eol + 2;
	}
	size_t body_have = n > hdr_len ? n - hdr_len : 0;
	size_t body_need = cl > 0 ? (size_t)cl : 0;
	if (body_need > 1024 * 1024) {
		*err = aram_net_error_new(EOVERFLOW, "HTTP body too large");
		return 0;
	}
	uint8_t *body = NULL;
	if (body_need > 0) {
		body = (uint8_t *)aram_alloc(body_need);
		if (body_have > 0) {
			size_t copy = body_have < body_need ? body_have : body_need;
			memcpy(body, buf + hdr_len, copy);
			body_have = copy;
		} else {
			body_have = 0;
		}
		while (body_have < body_need) {
			char sp;
			aram_gc_park_begin(&sp);
			ssize_t r = read((int)fd, body + body_have, body_need - body_have);
			aram_gc_park_end();
			if (r <= 0) {
				*err = r < 0 ? aram_net_errno_error() : aram_net_error_new(EINVAL, "truncated HTTP body");
				return 0;
			}
			body_have += (size_t)r;
		}
	} else if (body_have > 0) {
		body = (uint8_t *)aram_alloc(body_have);
		memcpy(body, buf + hdr_len, body_have);
		body_need = body_have;
	}
	out->method = method;
	out->path = path;
	out->body = (aram_slice_u8){ body, (int64_t)body_need, (int64_t)body_need };
	out->headers = headers;
	char *query = strchr((char *)out->path, '?');
	if (query) *query = '\0';
	return 1;
}

static void *aram_http_write(int64_t fd, int64_t code, const char *ctype, const uint8_t *data, int64_t len) {
	if (len < 0) return aram_net_error_new(EINVAL, "invalid HTTP body");
	if (!ctype) ctype = "application/octet-stream";
	const char *text = aram_http_status_text((int)code);
	char hdr[256];
	int hn = snprintf(hdr, sizeof hdr,
		"HTTP/1.0 %d %s\r\nContent-Type: %s\r\nContent-Length: %lld\r\nConnection: close\r\n\r\n",
		(int)code, text, ctype, (long long)len);
	if (hn < 0 || (size_t)hn >= sizeof hdr)
		return aram_net_error_new(EOVERFLOW, "HTTP response headers too large");
	size_t total = (size_t)hn + (size_t)len;
	uint8_t *msg = (uint8_t *)aram_alloc(total);
	memcpy(msg, hdr, (size_t)hn);
	if (len > 0) {
		if (!data) return aram_net_error_new(EINVAL, "invalid HTTP body");
		memcpy(msg + hn, data, (size_t)len);
	}
	aram_net_i64_result written = aram_net_write(fd, msg, (int64_t)total);
	return written.err;
}

static void *aram_http_write_str(int64_t fd, int64_t code, const char *ctype, const char *s) {
	size_t n = s ? strlen(s) : 0;
	return aram_http_write(fd, code, ctype, (const uint8_t *)s, (int64_t)n);
}

static void *aram_http_mux_dispatch(int64_t handle, aram_http_req req, int64_t fd) {
	aram_http_mux *mux = (aram_http_mux *)(intptr_t)handle;
	if (!mux) return aram_net_error_new(EINVAL, "invalid HTTP mux");
	for (int64_t i = 0; i < mux->len; i++) {
		if (strcmp(mux->routes[i].path, req.path) == 0) {
			aram_fn handler = mux->routes[i].handler;
			((aram_http_handler_fn)handler.fn)(handler.env, req, fd);
			return NULL;
		}
	}
	return aram_http_write_str(fd, 404, "text/plain; charset=utf-8", "காணவில்லை");
}

static void *aram_http_handle_conn(int64_t fd, aram_fn handler) {
	aram_http_req req;
	memset(&req, 0, sizeof req);
	void *err = NULL;
	if (!aram_http_read_req(fd, &req, &err)) {
		aram_http_write_str(fd, 400, "text/plain; charset=utf-8", "bad request");
		return err;
	}
	((aram_http_handler_fn)handler.fn)(handler.env, req, fd);
	return NULL;
}

static void *aram_http_handle_mux_conn(int64_t fd, int64_t mux) {
	aram_http_req req;
	memset(&req, 0, sizeof req);
	void *err = NULL;
	if (!aram_http_read_req(fd, &req, &err)) {
		aram_http_write_str(fd, 400, "text/plain; charset=utf-8", "bad request");
		return err;
	}
	return aram_http_mux_dispatch(mux, req, fd);
}

static void *aram_http_serve_one(int64_t ln, aram_fn handler) {
	aram_net_i64_result accepted = aram_net_accept(ln);
	if (accepted.err) return accepted.err;
	void *err = aram_http_handle_conn(accepted.value, handler);
	void *close_err = aram_net_close(accepted.value);
	return err ? err : close_err;
}

static void *aram_http_mux_serve_one(int64_t ln, int64_t mux) {
	aram_net_i64_result accepted = aram_net_accept(ln);
	if (accepted.err) return accepted.err;
	void *err = aram_http_handle_mux_conn(accepted.value, mux);
	void *close_err = aram_net_close(accepted.value);
	return err ? err : close_err;
}

typedef struct {
	aram_fn handler;
	int64_t fd;
} aram_http_job;

typedef struct {
	int64_t mux;
	int64_t fd;
} aram_http_mux_job;

static void *aram_http_job_main(void *p) {
	aram_thread_register();
	aram_http_job *j = (aram_http_job *)p;
	aram_http_handle_conn(j->fd, j->handler);
	aram_net_close(j->fd);
	free(j);
	aram_thread_unregister();
	return NULL;
}

static void *aram_http_mux_job_main(void *p) {
	aram_thread_register();
	aram_http_mux_job *j = (aram_http_mux_job *)p;
	aram_http_handle_mux_conn(j->fd, j->mux);
	aram_net_close(j->fd);
	free(j);
	aram_thread_unregister();
	return NULL;
}

static void *aram_http_listen_and_serve(const char *addr, aram_fn handler) {
	aram_net_i64_result listened = aram_net_listen(addr);
	if (listened.err) return listened.err;
	int64_t ln = listened.value;
	for (;;) {
		aram_gc_poll();
		aram_net_i64_result accepted = aram_net_accept(ln);
		if (accepted.err) return accepted.err;
		aram_http_job *j = (aram_http_job *)malloc(sizeof(*j));
		if (!j) return aram_net_error_new(ENOMEM, "HTTP worker allocation failed");
		j->handler = handler;
		j->fd = accepted.value;
		pthread_t th;
		if (pthread_create(&th, NULL, aram_http_job_main, j) != 0) {
			free(j);
			return aram_net_error_new(errno, strerror(errno));
		}
		pthread_detach(th);
	}
}

static void *aram_http_mux_listen_and_serve(const char *addr, int64_t mux) {
	if (!mux) return aram_net_error_new(EINVAL, "invalid HTTP mux");
	aram_net_i64_result listened = aram_net_listen(addr);
	if (listened.err) return listened.err;
	int64_t ln = listened.value;
	for (;;) {
		aram_gc_poll();
		aram_net_i64_result accepted = aram_net_accept(ln);
		if (accepted.err) return accepted.err;
		aram_http_mux_job *j = (aram_http_mux_job *)malloc(sizeof(*j));
		if (!j) return aram_net_error_new(ENOMEM, "HTTP worker allocation failed");
		j->mux = mux;
		j->fd = accepted.value;
		pthread_t th;
		if (pthread_create(&th, NULL, aram_http_mux_job_main, j) != 0) {
			free(j);
			return aram_net_error_new(errno, strerror(errno));
		}
		pthread_detach(th);
	}
}
