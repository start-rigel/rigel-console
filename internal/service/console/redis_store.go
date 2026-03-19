package console

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type redisSecurityStore struct {
	addr    string
	timeout time.Duration
}

func newRedisSecurityStore(addr string) (securityStore, error) {
	store := &redisSecurityStore{
		addr:    strings.TrimSpace(addr),
		timeout: 3 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()
	if err := store.ping(ctx); err != nil {
		return nil, fmt.Errorf("ping redis %s: %w", addr, err)
	}
	return store, nil
}

func NewRedisSecurityStore(addr string) (securityStore, error) {
	return newRedisSecurityStore(addr)
}

func (s *redisSecurityStore) LoadUsage(ctx context.Context, scope, key string) (usageState, bool, error) {
	var usage usageState
	ok, err := s.loadJSON(ctx, "usage:"+scope+":"+key, &usage)
	return usage, ok, err
}

func (s *redisSecurityStore) SaveUsage(ctx context.Context, scope, key string, usage usageState, ttl time.Duration) error {
	return s.saveJSON(ctx, "usage:"+scope+":"+key, usage, ttl)
}

func (s *redisSecurityStore) LoadRecommendation(ctx context.Context, key string) (storedRecommendation, bool, error) {
	var value storedRecommendation
	ok, err := s.loadJSON(ctx, "recommendation:"+key, &value)
	return value, ok, err
}

func (s *redisSecurityStore) SaveRecommendation(ctx context.Context, key string, value storedRecommendation, ttl time.Duration) error {
	return s.saveJSON(ctx, "recommendation:"+key, value, ttl)
}

func (s *redisSecurityStore) HasChallengePass(ctx context.Context, key string) (bool, error) {
	reply, err := s.execute(ctx, []string{"EXISTS", "challenge:" + key})
	if err != nil {
		return false, err
	}
	value, err := parseIntegerReply(reply)
	if err != nil {
		return false, err
	}
	return value > 0, nil
}

func (s *redisSecurityStore) SetChallengePass(ctx context.Context, key string, ttl time.Duration) error {
	_, err := s.execute(ctx, []string{"SET", "challenge:" + key, "1", "EX", strconv.Itoa(max(1, int(ttl.Seconds())))})
	return err
}

func (s *redisSecurityStore) LoadAdminSession(ctx context.Context, key string) (adminSession, bool, error) {
	var session adminSession
	ok, err := s.loadJSON(ctx, "admin_session:"+key, &session)
	return session, ok, err
}

func (s *redisSecurityStore) SaveAdminSession(ctx context.Context, key string, value adminSession, ttl time.Duration) error {
	return s.saveJSON(ctx, "admin_session:"+key, value, ttl)
}

func (s *redisSecurityStore) DeleteAdminSession(ctx context.Context, key string) error {
	_, err := s.execute(ctx, []string{"DEL", "admin_session:" + key})
	return err
}

func (s *redisSecurityStore) ping(ctx context.Context) error {
	reply, err := s.execute(ctx, []string{"PING"})
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(reply)) != "PONG" {
		return fmt.Errorf("unexpected ping reply %q", string(reply))
	}
	return nil
}

func (s *redisSecurityStore) saveJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.execute(ctx, []string{"SET", key, string(payload), "EX", strconv.Itoa(max(1, int(ttl.Seconds())))})
	return err
}

func (s *redisSecurityStore) loadJSON(ctx context.Context, key string, target any) (bool, error) {
	reply, err := s.execute(ctx, []string{"GET", key})
	if err != nil {
		if err == errRedisNil {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(reply, target); err != nil {
		return false, err
	}
	return true, nil
}

var errRedisNil = fmt.Errorf("redis nil")

func (s *redisSecurityStore) execute(ctx context.Context, args []string) ([]byte, error) {
	conn, err := s.dial(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(s.timeout))
	}

	if _, err := conn.Write(encodeRESP(args)); err != nil {
		return nil, err
	}
	return readRESP(bufio.NewReader(conn))
}

func (s *redisSecurityStore) dial(ctx context.Context) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: s.timeout}
	return dialer.DialContext(ctx, "tcp", s.addr)
}

func encodeRESP(args []string) []byte {
	var builder strings.Builder
	builder.WriteString("*")
	builder.WriteString(strconv.Itoa(len(args)))
	builder.WriteString("\r\n")
	for _, arg := range args {
		builder.WriteString("$")
		builder.WriteString(strconv.Itoa(len(arg)))
		builder.WriteString("\r\n")
		builder.WriteString(arg)
		builder.WriteString("\r\n")
	}
	return []byte(builder.String())
}

func readRESP(reader *bufio.Reader) ([]byte, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	switch prefix {
	case '+':
		return []byte(line), nil
	case '-':
		if strings.HasPrefix(strings.ToUpper(line), "NIL") {
			return nil, errRedisNil
		}
		return nil, fmt.Errorf("redis error: %s", line)
	case ':':
		return []byte(line), nil
	case '$':
		size, err := strconv.Atoi(line)
		if err != nil {
			return nil, err
		}
		if size < 0 {
			return nil, errRedisNil
		}
		payload := make([]byte, size+2)
		if _, err := reader.Read(payload); err != nil {
			return nil, err
		}
		return payload[:size], nil
	default:
		return nil, fmt.Errorf("unsupported redis reply prefix %q", prefix)
	}
}

func parseIntegerReply(reply []byte) (int, error) {
	return strconv.Atoi(strings.TrimSpace(string(reply)))
}
