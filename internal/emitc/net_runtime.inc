/* Aram Tamil-0.53: TCP sockets (Linux). Blocking I/O parks for GC. */
#include <arpa/inet.h>
#include <netdb.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>

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

static int64_t aram_net_listen(const char *addr) {
	char host[256];
	const char *port = NULL;
	if (aram_net_hostport(addr, host, sizeof host, &port) != 0) abort();
	struct addrinfo hints;
	memset(&hints, 0, sizeof hints);
	hints.ai_family = AF_INET;
	hints.ai_socktype = SOCK_STREAM;
	hints.ai_flags = AI_PASSIVE | AI_NUMERICSERV;
	const char *node = host[0] ? host : NULL;
	struct addrinfo *res = NULL;
	if (getaddrinfo(node, port, &hints, &res) != 0 || !res) abort();
	int fd = socket(res->ai_family, res->ai_socktype, res->ai_protocol);
	if (fd < 0) {
		freeaddrinfo(res);
		abort();
	}
	int yes = 1;
	setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof yes);
	if (bind(fd, res->ai_addr, res->ai_addrlen) != 0) {
		close(fd);
		freeaddrinfo(res);
		abort();
	}
	if (listen(fd, 16) != 0) {
		close(fd);
		freeaddrinfo(res);
		abort();
	}
	freeaddrinfo(res);
	return fd;
}

static int64_t aram_net_accept(int64_t ln) {
	char sp;
	aram_gc_park_begin(&sp);
	int fd = accept((int)ln, NULL, NULL);
	aram_gc_park_end();
	if (fd < 0) abort();
	return fd;
}

static int64_t aram_net_dial(const char *addr) {
	char host[256];
	const char *port = NULL;
	if (aram_net_hostport(addr, host, sizeof host, &port) != 0 || !host[0]) abort();
	struct addrinfo hints;
	memset(&hints, 0, sizeof hints);
	hints.ai_family = AF_INET;
	hints.ai_socktype = SOCK_STREAM;
	hints.ai_flags = AI_NUMERICSERV;
	struct addrinfo *res = NULL;
	if (getaddrinfo(host, port, &hints, &res) != 0 || !res) abort();
	int fd = socket(res->ai_family, res->ai_socktype, res->ai_protocol);
	if (fd < 0) {
		freeaddrinfo(res);
		abort();
	}
	char sp;
	aram_gc_park_begin(&sp);
	int rc = connect(fd, res->ai_addr, res->ai_addrlen);
	aram_gc_park_end();
	freeaddrinfo(res);
	if (rc != 0) {
		close(fd);
		abort();
	}
	return fd;
}

static int64_t aram_net_read(int64_t fd, uint8_t *data, int64_t len) {
	if (len < 0) abort();
	if (len == 0) return 0;
	if (!data) abort();
	char sp;
	aram_gc_park_begin(&sp);
	ssize_t n = read((int)fd, data, (size_t)len);
	aram_gc_park_end();
	if (n < 0) abort();
	return (int64_t)n;
}

static int64_t aram_net_write(int64_t fd, uint8_t *data, int64_t len) {
	if (len < 0) abort();
	if (len == 0) return 0;
	if (!data) abort();
	int64_t off = 0;
	while (off < len) {
		char sp;
		aram_gc_park_begin(&sp);
		ssize_t n = write((int)fd, data + off, (size_t)(len - off));
		aram_gc_park_end();
		if (n <= 0) abort();
		off += n;
	}
	return off;
}

static void aram_net_close(int64_t fd) {
	if (fd >= 0) close((int)fd);
}

static const char *aram_net_addr(int64_t fd) {
	struct sockaddr_in sa;
	socklen_t slen = sizeof sa;
	memset(&sa, 0, sizeof sa);
	if (getsockname((int)fd, (struct sockaddr *)&sa, &slen) != 0) abort();
	char ip[INET_ADDRSTRLEN];
	if (!inet_ntop(AF_INET, &sa.sin_addr, ip, sizeof ip)) abort();
	if (strcmp(ip, "0.0.0.0") == 0) memcpy(ip, "127.0.0.1", 10);
	unsigned port = ntohs(sa.sin_port);
	char buf[64];
	int n = snprintf(buf, sizeof buf, "%s:%u", ip, port);
	if (n < 0 || (size_t)n >= sizeof buf) abort();
	char *r = (char *)aram_alloc((size_t)n + 1);
	memcpy(r, buf, (size_t)n + 1);
	return r;
}
