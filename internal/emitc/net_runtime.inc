/* Aram Tamil-0.55: TCP sockets (Linux). Blocking I/O parks for GC. */
#include <arpa/inet.h>
#include <errno.h>
#include <netdb.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>

typedef struct { int64_t value; void *err; } aram_net_i64_result;
typedef struct { const char *value; void *err; } aram_net_str_result;
typedef struct { int64_t n; const char *addr; void *err; } aram_net_udp_result;
typedef struct { int64_t code; const char *message; } aram_net_error;

static void *aram_net_error_new(int64_t code, const char *message) {
	aram_net_error *err = (aram_net_error *)aram_alloc(sizeof(*err));
	err->code = code;
	err->message = message ? message : "network error";
	return err;
}

static void *aram_net_errno_error(void) {
	return aram_net_error_new(errno, strerror(errno));
}

static int aram_net_hostport(const char *addr, char *host, size_t hostcap, const char **port) {
	if (!addr || !host || !port || hostcap < 2) return -1;
	const char *colon = strrchr(addr, ':');
	if (!colon || colon[1] == '\0') return -1;
	*port = colon + 1;
	size_t hl = (size_t)(colon - addr);
	if (hl >= hostcap) return -1;
	if (hl == 0) {
		host[0] = '\0';
		return 0;
	}
	memcpy(host, addr, hl);
	host[hl] = '\0';
	return 0;
}

static aram_net_i64_result aram_net_listen(const char *addr) {
	char host[256];
	const char *port = NULL;
	if (aram_net_hostport(addr, host, sizeof host, &port) != 0)
		return (aram_net_i64_result){ 0, aram_net_error_new(EINVAL, "invalid listen address") };
	struct addrinfo hints;
	memset(&hints, 0, sizeof hints);
	hints.ai_family = AF_INET;
	hints.ai_socktype = SOCK_STREAM;
	hints.ai_flags = AI_PASSIVE | AI_NUMERICSERV;
	const char *node = host[0] ? host : NULL;
	struct addrinfo *res = NULL;
	int gai = getaddrinfo(node, port, &hints, &res);
	if (gai != 0 || !res)
		return (aram_net_i64_result){ 0, aram_net_error_new(gai, gai_strerror(gai)) };
	int fd = socket(res->ai_family, res->ai_socktype, res->ai_protocol);
	if (fd < 0) {
		freeaddrinfo(res);
		return (aram_net_i64_result){ 0, aram_net_errno_error() };
	}
	int yes = 1;
	setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof yes);
	if (bind(fd, res->ai_addr, res->ai_addrlen) != 0) {
		close(fd);
		freeaddrinfo(res);
		return (aram_net_i64_result){ 0, aram_net_errno_error() };
	}
	if (listen(fd, 16) != 0) {
		close(fd);
		freeaddrinfo(res);
		return (aram_net_i64_result){ 0, aram_net_errno_error() };
	}
	freeaddrinfo(res);
	return (aram_net_i64_result){ fd, NULL };
}

static aram_net_i64_result aram_net_accept(int64_t ln) {
	char sp;
	aram_gc_park_begin(&sp);
	int fd = accept((int)ln, NULL, NULL);
	aram_gc_park_end();
	if (fd < 0) return (aram_net_i64_result){ 0, aram_net_errno_error() };
	return (aram_net_i64_result){ fd, NULL };
}

static aram_net_i64_result aram_net_dial(const char *addr) {
	char host[256];
	const char *port = NULL;
	if (aram_net_hostport(addr, host, sizeof host, &port) != 0 || !host[0])
		return (aram_net_i64_result){ 0, aram_net_error_new(EINVAL, "invalid dial address") };
	struct addrinfo hints;
	memset(&hints, 0, sizeof hints);
	hints.ai_family = AF_INET;
	hints.ai_socktype = SOCK_STREAM;
	hints.ai_flags = AI_NUMERICSERV;
	struct addrinfo *res = NULL;
	int gai = getaddrinfo(host, port, &hints, &res);
	if (gai != 0 || !res)
		return (aram_net_i64_result){ 0, aram_net_error_new(gai, gai_strerror(gai)) };
	int fd = socket(res->ai_family, res->ai_socktype, res->ai_protocol);
	if (fd < 0) {
		freeaddrinfo(res);
		return (aram_net_i64_result){ 0, aram_net_errno_error() };
	}
	char sp;
	aram_gc_park_begin(&sp);
	int rc = connect(fd, res->ai_addr, res->ai_addrlen);
	aram_gc_park_end();
	freeaddrinfo(res);
	if (rc != 0) {
		close(fd);
		return (aram_net_i64_result){ 0, aram_net_errno_error() };
	}
	return (aram_net_i64_result){ fd, NULL };
}

static aram_net_i64_result aram_net_read(int64_t fd, uint8_t *data, int64_t len) {
	if (len < 0) return (aram_net_i64_result){ 0, aram_net_error_new(EINVAL, "invalid read buffer") };
	if (len == 0) return (aram_net_i64_result){ 0, NULL };
	if (!data) return (aram_net_i64_result){ 0, aram_net_error_new(EINVAL, "invalid read buffer") };
	char sp;
	aram_gc_park_begin(&sp);
	ssize_t n = read((int)fd, data, (size_t)len);
	aram_gc_park_end();
	if (n < 0) return (aram_net_i64_result){ 0, aram_net_errno_error() };
	return (aram_net_i64_result){ (int64_t)n, NULL };
}

static aram_net_i64_result aram_net_write(int64_t fd, const uint8_t *data, int64_t len) {
	if (len < 0) return (aram_net_i64_result){ 0, aram_net_error_new(EINVAL, "invalid write buffer") };
	if (len == 0) return (aram_net_i64_result){ 0, NULL };
	if (!data) return (aram_net_i64_result){ 0, aram_net_error_new(EINVAL, "invalid write buffer") };
	int64_t off = 0;
	while (off < len) {
		char sp;
		aram_gc_park_begin(&sp);
		ssize_t n = write((int)fd, data + off, (size_t)(len - off));
		aram_gc_park_end();
		if (n <= 0) return (aram_net_i64_result){ off, aram_net_errno_error() };
		off += n;
	}
	return (aram_net_i64_result){ off, NULL };
}

static void *aram_net_close(int64_t fd) {
	if (fd >= 0 && close((int)fd) != 0) return aram_net_errno_error();
	return NULL;
}

static aram_net_str_result aram_net_addr(int64_t fd) {
	struct sockaddr_in sa;
	socklen_t slen = sizeof sa;
	memset(&sa, 0, sizeof sa);
	if (getsockname((int)fd, (struct sockaddr *)&sa, &slen) != 0)
		return (aram_net_str_result){ "", aram_net_errno_error() };
	char ip[INET_ADDRSTRLEN];
	if (!inet_ntop(AF_INET, &sa.sin_addr, ip, sizeof ip))
		return (aram_net_str_result){ "", aram_net_errno_error() };
	if (strcmp(ip, "0.0.0.0") == 0) memcpy(ip, "127.0.0.1", 10);
	unsigned port = ntohs(sa.sin_port);
	char buf[64];
	int n = snprintf(buf, sizeof buf, "%s:%u", ip, port);
	if (n < 0 || (size_t)n >= sizeof buf)
		return (aram_net_str_result){ "", aram_net_error_new(EINVAL, "address formatting failed") };
	char *r = (char *)aram_alloc((size_t)n + 1);
	memcpy(r, buf, (size_t)n + 1);
	return (aram_net_str_result){ r, NULL };
}

static aram_net_i64_result aram_net_udp_listen(const char *addr) {
	char host[256];
	const char *port = NULL;
	if (aram_net_hostport(addr, host, sizeof host, &port) != 0)
		return (aram_net_i64_result){ 0, aram_net_error_new(EINVAL, "invalid UDP listen address") };
	struct addrinfo hints;
	memset(&hints, 0, sizeof hints);
	hints.ai_family = AF_INET;
	hints.ai_socktype = SOCK_DGRAM;
	hints.ai_flags = AI_PASSIVE | AI_NUMERICSERV;
	struct addrinfo *res = NULL;
	int gai = getaddrinfo(host[0] ? host : NULL, port, &hints, &res);
	if (gai != 0 || !res) return (aram_net_i64_result){ 0, aram_net_error_new(gai, gai_strerror(gai)) };
	int fd = socket(res->ai_family, res->ai_socktype, res->ai_protocol);
	if (fd < 0) { freeaddrinfo(res); return (aram_net_i64_result){ 0, aram_net_errno_error() }; }
	if (bind(fd, res->ai_addr, res->ai_addrlen) != 0) {
		close(fd); freeaddrinfo(res); return (aram_net_i64_result){ 0, aram_net_errno_error() };
	}
	freeaddrinfo(res);
	return (aram_net_i64_result){ fd, NULL };
}

static aram_net_i64_result aram_net_udp_send(int64_t fd, const char *addr, const uint8_t *data, int64_t len) {
	char host[256];
	const char *port = NULL;
	if (len < 0 || (len > 0 && !data)) return (aram_net_i64_result){ 0, aram_net_error_new(EINVAL, "invalid UDP data") };
	if (aram_net_hostport(addr, host, sizeof host, &port) != 0 || !host[0])
		return (aram_net_i64_result){ 0, aram_net_error_new(EINVAL, "invalid UDP address") };
	struct addrinfo hints;
	memset(&hints, 0, sizeof hints);
	hints.ai_family = AF_INET;
	hints.ai_socktype = SOCK_DGRAM;
	hints.ai_flags = AI_NUMERICSERV;
	struct addrinfo *res = NULL;
	int gai = getaddrinfo(host, port, &hints, &res);
	if (gai != 0 || !res) return (aram_net_i64_result){ 0, aram_net_error_new(gai, gai_strerror(gai)) };
	volatile const uint8_t *root = data;
	char sp;
	aram_gc_park_begin(&sp);
	ssize_t n = sendto((int)fd, (const uint8_t *)root, (size_t)len, 0, res->ai_addr, res->ai_addrlen);
	aram_gc_park_end();
	freeaddrinfo(res);
	if (n < 0) return (aram_net_i64_result){ 0, aram_net_errno_error() };
	return (aram_net_i64_result){ (int64_t)n, NULL };
}

static aram_net_udp_result aram_net_udp_recv(int64_t fd, uint8_t *data, int64_t len) {
	if (len < 0 || (len > 0 && !data)) return (aram_net_udp_result){ 0, "", aram_net_error_new(EINVAL, "invalid UDP buffer") };
	struct sockaddr_in peer;
	socklen_t plen = sizeof peer;
	volatile uint8_t *root = data;
	char sp;
	aram_gc_park_begin(&sp);
	ssize_t n = recvfrom((int)fd, (uint8_t *)root, (size_t)len, 0, (struct sockaddr *)&peer, &plen);
	aram_gc_park_end();
	if (n < 0) return (aram_net_udp_result){ 0, "", aram_net_errno_error() };
	char ip[INET_ADDRSTRLEN];
	if (!inet_ntop(AF_INET, &peer.sin_addr, ip, sizeof ip)) return (aram_net_udp_result){ 0, "", aram_net_errno_error() };
	char buf[64];
	int m = snprintf(buf, sizeof buf, "%s:%u", ip, ntohs(peer.sin_port));
	if (m < 0 || (size_t)m >= sizeof buf) return (aram_net_udp_result){ 0, "", aram_net_error_new(EINVAL, "address formatting failed") };
	char *out = (char *)aram_alloc((size_t)m + 1);
	memcpy(out, buf, (size_t)m + 1);
	return (aram_net_udp_result){ (int64_t)n, out, NULL };
}
